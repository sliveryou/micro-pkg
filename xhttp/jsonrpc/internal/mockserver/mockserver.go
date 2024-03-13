package mockserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/sliveryou/go-tool/v2/convert"
	"github.com/sliveryou/go-tool/v2/id-generator/uuid"
	"github.com/sliveryou/go-tool/v2/randx"
)

// mock server constants
const (
	DefaultErrCode = 99

	MethodGetInt64     = "mock.GetInt64"
	MethodGetFloat64   = "mock.GetFloat64"
	MethodGetBool      = "mock.GetBool"
	MethodGetString    = "mock.GetString"
	MethodReadToObject = "mock.ReadToObject"

	MethodGetStdErr    = "mock.GetStdErr"
	MethodGetStringErr = "mock.GetStringErr"

	ExtensionRawRequest = "raw_request"
)

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error error method
func (e *rpcError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type mockServer struct {
	server http.Server
}

// JSONRPCHandler json rpc single handler
func (s *mockServer) JSONRPCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, nil, errors.Errorf("%s http method not allowed", r.Method))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, nil, errors.WithMessage(err, "read all body err"))
		return
	}

	req := make(map[string]any)
	err = unmarshal(body, &req)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, nil, errors.WithMessage(err, "json unmarshal err"))
		return
	}

	id := req["id"]
	result, code, err := handleReq(string(body), req)
	if err != nil {
		writeErr(w, code, id, err)
		return
	}

	writeOK(w, id, result, string(body))
}

// JSONRPCBatchHandler json rpc batch handler
func (s *mockServer) JSONRPCBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, nil, errors.Errorf("%s method not allowed", r.Method))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, nil, errors.WithMessage(err, "read all body err"))
		return
	}

	var reqs []map[string]any
	err = unmarshal(body, &reqs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, nil, errors.WithMessage(err, "json unmarshal err"))
		return
	}

	var resps []map[string]any
	for _, req := range reqs {
		id := req["id"]
		result, _, err := handleReq(string(body), req)
		if err != nil {
			resps = append(resps, getErrResp(id, err))
		} else {
			resps = append(resps, getOKResp(id, result, ""))
		}
	}

	writeOKRaw(w, resps)
}

func handleReq(rawRequest string, req map[string]any) (result any, code int, err error) {
	_, ok := req["method"]
	if !ok {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	method, ok := req["method"].(string)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("invalid method type")
	}

	switch method {
	case MethodGetInt64:
		result = convert.ToInt64(randx.NewNumber(6))
	case MethodGetFloat64:
		result = convert.ToFloat64(randx.NewNumber(3) + "." + randx.NewNumber(3))
	case MethodGetBool:
		result = time.Now().Unix()%2 == 0
	case MethodGetString:
		result = randx.NewString(12)
	case MethodReadToObject:
		obj := map[string]any{
			"name":    "sliveryou",
			"age":     18,
			"country": "China",
			"uuid":    uuid.NextV4(),
		}
		if rawRequest != "" {
			obj[ExtensionRawRequest] = rawRequest
		}
		result = obj
	case MethodGetStdErr:
		return nil, http.StatusInternalServerError, &rpcError{
			Code:    DefaultErrCode,
			Message: "method get std err",
			Data:    "method get std err",
		}
	case MethodGetStringErr:
		return nil, http.StatusInternalServerError, &rpcError{
			Message: "method get string err",
		}
	default:
		return nil, http.StatusInternalServerError, errors.New("method not found")
	}

	return result, http.StatusOK, nil
}

func getErrResp(id any, err error) map[string]any {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
	}

	var returnErr any
	var rpcErr *rpcError
	if errors.As(err, &rpcErr) {
		if rpcErr.Code == 0 && rpcErr.Data == nil {
			returnErr = rpcErr.Message
		} else {
			returnErr = rpcErr
		}
	} else {
		returnErr = map[string]any{
			"code":    DefaultErrCode,
			"message": errors.Cause(err).Error(),
			"data":    err.Error(),
		}
	}

	resp["error"] = returnErr

	return resp
}

func getOKResp(id, result any, rawRequest string) map[string]any {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	if rawRequest != "" {
		resp[ExtensionRawRequest] = rawRequest
	}

	return resp
}

func writeErr(w http.ResponseWriter, code int, id any, err error) {
	httpx.WriteJson(w, code, getErrResp(id, err))
}

func writeOK(w http.ResponseWriter, id, result any, rawRequest string) {
	httpx.WriteJson(w, http.StatusOK, getOKResp(id, result, rawRequest))
}

func writeOKRaw(w http.ResponseWriter, raw any) {
	httpx.WriteJson(w, http.StatusOK, raw)
}

func unmarshal(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()

	return d.Decode(v)
}

var server *mockServer

func initServer(addr ...string) {
	server = &mockServer{}
	mux := http.NewServeMux()
	mux.Handle("/jsonrpc", http.HandlerFunc(server.JSONRPCHandler))
	mux.Handle("/jsonrpc/batch", http.HandlerFunc(server.JSONRPCBatchHandler))
	server.server.Handler = mux
	server.server.Addr = ":18090"
	if len(addr) > 0 {
		server.server.Addr = addr[0]
	}
}

// Run mock server
func Run(addr ...string) error {
	initServer(addr...)
	return server.server.ListenAndServe()
}

// Close mock server
func Close() error {
	return server.server.Shutdown(context.TODO())
}
