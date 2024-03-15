package face

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/collection"

	"github.com/sliveryou/go-tool/v2/validator"

	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/xhttp"
)

// 百度云人脸识别认证 API：
//  鉴权认证 token 获取 API：https://ai.baidu.com/ai-doc/REFERENCE/Ck3dwjhhu
//  视频活体检测 API：https://ai.baidu.com/ai-doc/FACE/lk37c1tag
//  人脸实名认证 API：https://ai.baidu.com/ai-doc/FACE/7k37c1ucj
//  错误码：https://ai.baidu.com/ai-doc/FACE/5k37c1ujz

const (
	// AccessTokenURL 鉴权认证 token 获取接口地址
	AccessTokenURL = "https://aip.baidubce.com/oauth/2.0/token"
	// VideoVerifyURL 视频活体检测接口地址
	VideoVerifyURL = "https://aip.baidubce.com/rest/2.0/face/v1/faceliveness/verify"
	// PersonVerifyURL 人脸实名认证接口地址
	PersonVerifyURL = "https://aip.baidubce.com/rest/2.0/face/v3/person/verify"

	cacheName = "face auth"
	cacheKey  = "access token"

	defaultGrantType = "client_credentials"
	defaultImageType = "BASE64"
)

var (
	// PersonVerifyThreshold 与公民身份证小图相似度可能性阈值，超过即判断为同一人
	PersonVerifyThreshold = 80.0

	// ErrFaceVideoVerify 视频活体检测错误
	ErrFaceVideoVerify = errcode.NewCommon("活体检测失败")
	// ErrFacePersonVerify 人脸实名认证错误
	ErrFacePersonVerify = errcode.NewCommon("人脸与公民身份证小图相似度匹配过低")
)

// Config 人脸识别认证相关配置
type Config struct {
	IsMock    bool   `json:",optional"` // 是否模拟通过
	APIKey    string `json:",optional"` // 接口Key
	SecretKey string `json:",optional"` // 接口密钥
}

// Face 人脸识别认证器结构详情
type Face struct {
	c      Config
	client *xhttp.Client
	cache  *collection.Cache
}

// NewFace 新建人脸识别认证器
func NewFace(c Config) (*Face, error) {
	if !c.IsMock {
		if c.APIKey == "" || c.SecretKey == "" {
			return nil, errors.New("face: illegal face config")
		}
	}

	// 设置鉴权认证 token 缓存，过期时间为 24 小时
	cache, err := collection.NewCache(24*time.Hour, collection.WithName(cacheName))
	if err != nil {
		return nil, errors.WithMessage(err, "face: new cache err")
	}

	cc := xhttp.DefaultConfig()
	cc.HTTPTimeout = 30 * time.Second

	return &Face{c: c, client: xhttp.NewClient(cc), cache: cache}, nil
}

// MustNewFace 新建人脸识别认证器
func MustNewFace(c Config) *Face {
	f, err := NewFace(c)
	if err != nil {
		panic(err)
	}

	return f
}

// AuthenticateRequest 人脸识别认证请求
type AuthenticateRequest struct {
	Name        string `validate:"required" label:"姓名"`          // 姓名
	IDCard      string `validate:"required,idcard" label:"身份证号"` // 身份证号
	VideoBase64 string `validate:"required,base64" label:"视频数据"` // base64 编码的视频数据（建议视频大小控制在 10MB/1min 以内）
}

// AuthenticateResponse 人脸识别认证响应
type AuthenticateResponse struct {
	LogID int64 // 日志ID
}

// Authenticate 人脸识别认证
func (f *Face) Authenticate(ctx context.Context, req *AuthenticateRequest) (*AuthenticateResponse, error) {
	if f.c.IsMock {
		return &AuthenticateResponse{}, nil
	}

	s := strings.SplitN(req.VideoBase64, ",", 2)
	if len(s) <= 1 {
		req.VideoBase64 = s[0]
	} else {
		req.VideoBase64 = s[1]
	}

	// 校验请求参数
	if err := validator.Verify(req); err != nil {
		return nil, errcode.New(errcode.CodeInvalidParams, err.Error())
	}

	// 获取鉴权认证 token
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get access token err")
	}

	// 视频活体检测
	validPic, err := f.videoVerify(ctx, accessToken, req.VideoBase64)
	if err != nil {
		return nil, errors.WithMessage(err, "video verify err")
	}

	// 人脸实名认证
	logID, err := f.personVerify(ctx, accessToken, validPic, req.Name, req.IDCard)
	if err != nil {
		return nil, errors.WithMessage(err, "person verify err")
	}

	return &AuthenticateResponse{LogID: logID}, nil
}

// getAccessToken 获取鉴权认证 token
func (f *Face) getAccessToken(ctx context.Context) (string, error) {
	item, err := f.cache.Take(cacheKey, func() (any, error) {
		var resp getAccessTokenResp

		rawURL := AccessTokenURL
		values := make(url.Values)
		values.Set("grant_type", defaultGrantType)
		values.Set("client_id", f.c.APIKey)
		values.Set("client_secret", f.c.SecretKey)
		rawURL += "?" + values.Encode()

		_, err := f.client.Call(ctx, http.MethodGet, rawURL, nil, nil, &resp)
		if err != nil {
			return "", errors.WithMessage(err, "client call err")
		}

		if resp.AccessToken == "" {
			return "", errors.New(resp.Error + ": " + resp.ErrorDescription)
		}

		return resp.AccessToken, nil
	})
	if err != nil {
		return "", errors.WithMessage(err, "cache take err")
	}
	accessToken, ok := item.(string)
	if !ok {
		return "", errors.New("cache take access token err")
	}

	return accessToken, nil
}

// videoVerify 视频活体检测
func (f *Face) videoVerify(ctx context.Context, accessToken, videoBase64 string) (string, error) {
	var resp videoVerifyResp

	rawURL := VideoVerifyURL + "?access_token=" + accessToken
	header := map[string]string{
		xhttp.HeaderContentType: xhttp.MIMEForm,
	}
	data := "video_base64=" + videoBase64

	_, err := f.client.Call(ctx, http.MethodPost, rawURL, header, strings.NewReader(data), &resp)
	if err != nil {
		return "", errors.WithMessage(err, "client call err")
	}

	if resp.ErrorCode != codeSuccess {
		if errMsg, ok := errMap[resp.ErrorCode]; ok {
			return "", errcode.NewCommon(errMsg)
		}
		return "", errors.Errorf("%d: %s", resp.ErrorCode, resp.ErrorMsg)
	}

	var result videoVerifyResult
	err = json.Unmarshal(resp.Result, &result)
	if err != nil {
		return "", errors.WithMessage(err, "json unmarshal result err")
	}

	var validPic string
	for _, pic := range result.PicList {
		// 任何一张通过了阈值，即可判断为活体，建议可用三帧情况
		if pic.LivenessScore >= result.Thresholds.Free1e3 {
			validPic = pic.Pic
			break
		}
	}

	if validPic == "" {
		return "", ErrFaceVideoVerify
	}

	return validPic, nil
}

// personVerify 人脸实名认证
func (f *Face) personVerify(ctx context.Context, accessToken, validPic, name, idCard string) (int64, error) {
	var resp personVerifyResp

	rawURL := fmt.Sprintf("%s?access_token=%s", PersonVerifyURL, accessToken)
	header := map[string]string{
		xhttp.HeaderContentType: xhttp.MIMEApplicationJSON,
	}
	req := &personVerifyReq{
		Image:        validPic,
		ImageType:    defaultImageType,
		IDCardNumber: idCard,
		Name:         name,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return 0, errors.WithMessage(err, "json marshal request err")
	}

	_, err = f.client.Call(ctx, http.MethodPost, rawURL, header, bytes.NewReader(b), &resp)
	if err != nil {
		return 0, errors.WithMessage(err, "client call err")
	}

	if resp.ErrorCode != codeSuccess {
		if errMsg, ok := errMap[resp.ErrorCode]; ok {
			return 0, errcode.NewCommon(errMsg)
		}
		return 0, errors.Errorf("%d: %s", resp.ErrorCode, resp.ErrorMsg)
	}

	if resp.Result.Score < PersonVerifyThreshold {
		return 0, ErrFacePersonVerify
	}

	return resp.LogID, nil
}

const (
	// codeSuccess 接口请求成功状态码
	codeSuccess = 0
)

// getAccessTokenResp 获取鉴权认证 token 响应
type getAccessTokenResp struct {
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	Scope            string `json:"scope"`
	SessionKey       string `json:"session_key"`
	AccessToken      string `json:"access_token"`
	SessionSecret    string `json:"session_secret"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// videoVerifyResp 视频活体检测响应
type videoVerifyResp struct {
	ErrNo       int64           `json:"err_no"`
	ErrMsg      string          `json:"err_msg"`
	Result      json.RawMessage `json:"result"`
	Timestamp   int64           `json:"timestamp"`
	Cached      int64           `json:"cached"`
	ServerLogID int64           `json:"serverlogid"`
	ErrorCode   int64           `json:"error_code"`
	ErrorMsg    string          `json:"error_msg"`
}

// videoVerifyResult 视频活体检测结果
type videoVerifyResult struct {
	PicList     []videoVerifyPic      `json:"pic_list"`              // 1-8 张抽取出来的图片信息（默认返回 8 张图片）
	Score       float64               `json:"score"`                 // 活体检测的总体打分，范围 [0,1]，分数越高则活体的概率越大
	MaxSpoofing float64               `json:"maxpoofing"`            // 判断是否是合成图功能，范围 [0,1]，分数越高则概率越大
	Thresholds  videoVerifyThresholds `json:"videoVerifyThresholds"` // 阈值，按活体检测分数 > 阈值来判定活体检测是否通过（阈值视产品需求选择其中一个）
}

// videoVerifyPic 视频活体检测图片信息
type videoVerifyPic struct {
	Pic           string  `json:"videoVerifyPic"` // base64 编码后的图片信息
	FaceToken     string  `json:"face_token"`     // 人脸图片的唯一标识
	FaceID        string  `json:"face_id"`        // 人脸图片ID
	LivenessScore float64 `json:"liveness_score"` // 此图片的活体分数，范围 [0,1]
	Spoofing      float64 `json:"spoofing"`       // 判断此图片是合成图的分数，范围 [0,1]
}

// videoVerifyThresholds 视频活体检测阈值
type videoVerifyThresholds struct {
	Free1e4 float64 `json:"frr_1e-4"`
	Free1e3 float64 `json:"frr_1e-3"`
	Free1e2 float64 `json:"frr_1e-2"`
}

// personVerifyReq 人脸实名认证请求
type personVerifyReq struct {
	Image        string `json:"image"`          // 图片信息（总数据大小应小于 10MB）
	ImageType    string `json:"image_type"`     // 图片类型，BASE64：图片的 base64 值，base64 编码后的图片数据，编码后的图片大小不超过 2MB；图片尺寸不超过 1920*1080
	IDCardNumber string `json:"id_card_number"` // 身份证号码（注：需要是 UTF-8 编码的中文）
	Name         string `json:"name"`           // 姓名
}

// personVerifyResp 人脸实名认证响应
type personVerifyResp struct {
	ErrorCode int64              `json:"error_code"` // 错误码
	ErrorMsg  string             `json:"error_msg"`  // 错误消息
	LogID     int64              `json:"log_id"`     // 日志ID
	Result    personVerifyResult `json:"result"`     // 结果
}

// personVerifyResult 人脸实名认证结果
type personVerifyResult struct {
	Score float64 `json:"score"` // 与公民身份证小图相似度可能性，用于验证生活照与公民身份证小图是否为同一人，有正常分数时为 [0~100]，推荐阈值 80，超过即判断为同一人
}

var errMap = map[int64]string{
	// H5 活体检测接口错误码：
	// https://ai.baidu.com/ai-doc/FACE/5k37c1ujz#h5%E6%B4%BB%E4%BD%93%E6%A3%80%E6%B5%8B%E6%8E%A5%E5%8F%A3%E9%94%99%E8%AF%AF%E7%A0%81%E5%88%97%E8%A1%A8
	216432: "视频解析服务调用失败，请重新尝试",
	216433: "视频解析服务发生错误，请重新尝试",
	216501: "没有找到人脸，请查看上传视频是否包含人脸",
	216509: "视频中的声音无法识别，请重新录制视频",
	216510: "动作活体模式验证时视频长度超过10s，请重新录制时长小于10s的视频",
	216908: "视频中人脸质量过低，请重新录制视频",
	216612: "系统繁忙，请重新尝试",

	// 人脸实名认证 H5 方案错误码：
	// https://ai.baidu.com/ai-doc/FACE/5k37c1ujz#%E4%BA%BA%E8%84%B8%E5%AE%9E%E5%90%8D%E8%AE%A4%E8%AF%81h5%E6%96%B9%E6%A1%88%E9%94%99%E8%AF%AF%E7%A0%81%E5%88%97%E8%A1%A8
	283456: "图片为空或格式不正确",
	283460: "视频文件过大，核验请求超时",
	222202: "图片中没有人脸",
	222203: "无法解析人脸",
	223113: "人脸有被遮挡",
	223114: "人脸模糊",
	223115: "人脸光照不好",
	223116: "人脸不完整",
	223129: "人脸未面向正前方",
	223131: "检测到图片为合成图，不符合要求",
	216434: "活体检测未通过，可能存在欺诈行为",
	216508: "视频中有多张人脸",
	222350: "公安网图片不存在或质量过低，请将此次身份验证转到人工进行处理",
	222351: "身份证号与姓名不匹配",
	222022: "身份证号不符合格式要求",
	222023: "姓名格式错误",
	222354: "公安库里不存在此身份证号",
	222355: "身份证号码正确，公安库里没有对应的照片",
	222356: "验证的人脸图片质量不符合要求",
	222360: "身份核验未通过，无法确认身份",
	222361: "公安服务连接失败",
	216600: "输入身份证格式错误",
	216601: "身份证号和名字不匹配",
}
