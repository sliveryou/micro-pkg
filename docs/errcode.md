# 接口错误码

> 注意：HTTP 状态码为 `200`，错误码为 `0` 时，才代表接口请求成功

| **错误** | **错误码** | **释义** | **HTTP 状态码** |
|:---------|:-----------|:---------|:----------------|
| OK | 0 | ok | <font color='green'>200</font> |
| ErrCommon | 97 | 通用错误 | <font color='green'>200</font> |
| ErrRecordNotFound | 98 | 记录不存在 | <font color='green'>200</font> |
| ErrUnexpected | 99 | 服务器繁忙，请稍后重试 | <font color='green'>200</font> |
| ErrInvalidParams | 100 | 请求参数错误 | <font color='green'>200</font> |
| ErrInvalidSignParams | 110 | 签名参数错误 | <font color='green'>200</font> |
| ErrInvalidContentMD5 | 111 | Content-MD5 错误 | <font color='green'>200</font> |
| ErrBodyTooLarge | 112 | 请求体过大 | <font color='green'>200</font> |
| ErrBankCard4CAuth | 115 | 银行卡四要素认证失败 | <font color='green'>200</font> |
| ErrCorpAccountAuth | 116 | 企业银行卡账户认证失败 | <font color='green'>200</font> |
| ErrFaceVideoAuth | 117 | 视频活体检测失败 | <font color='green'>200</font> |
| ErrFacePersonAuth | 118 | 人脸与公民身份证小图相似度匹配过低 | <font color='green'>200</font> |
| ErrGetExpressFailed | 120 | 查询快递失败 | <font color='green'>200</font> |
| ErrProviderOverQuota | 130 | 该提供方今日请求超过配额，请明日再试 | <font color='green'>200</font> |
| ErrIPSourceOverQuota | 131 | 该 IP 地址今日请求超过配额，请明日再试 | <font color='green'>200</font> |
| ErrReceiverOverQuota | 132 | 该接收方今日请求超过配额，请明日再试 | <font color='green'>200</font> |
| ErrSendTooFrequently | 133 | 发送过于频繁，请稍后再试 | <font color='green'>200</font> |
| ErrVerifyTooFrequently | 134 | 验证过于频繁，请稍后再试 | <font color='green'>200</font> |
| ErrEmailUnsupported | 135 | 暂不支持邮件通知服务 | <font color='green'>200</font> |
| ErrSmsUnsupported | 136 | 暂不支持短信通知服务 | <font color='green'>200</font> |
| ErrEmailTmplNotFound | 137 | 邮件模板信息不存在 | <font color='green'>200</font> |
| ErrSmsTmplNotFound | 138 | 短信模板信息不存在 | <font color='green'>200</font> |
| ErrInvalidCaptcha | 139 | 验证码错误 | <font color='green'>200</font> |
| ErrCaptchaNotFound | 140 | 验证码不存在或已过期 | <font color='green'>200</font> |
| ErrInvalidSign | 150 | 签名错误 | <font color='red'>401</font> |
| ErrSignExpired | 151 | 签名已过期 | <font color='red'>401</font> |
| ErrNonceExpired | 152 | 随机数已过期 | <font color='red'>401</font> |
| ErrInvalidToken | 153 | Token 错误 | <font color='red'>401</font> |
| ErrAPINotAllowed | 154 | 暂不支持该 API | <font color='green'>200</font> |
| ErrRPCNotAllowed | 155 | 暂不支持该 RPC | <font color='green'>200</font> |
