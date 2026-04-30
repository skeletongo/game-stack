package stack

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
)

// Response 统一响应结构。
// 所有客户端消息的回复都使用此结构包裹。
type Response struct {
	Code    int32  `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK() *Response {
	return &Response{Code: 0, Message: "ok"}
}

// OKWithData 返回带数据的成功响应。
func OKWithData(data any) *Response {
	return &Response{Code: 0, Message: "ok", Data: data}
}

// Err 返回错误响应。
func Err(c *Code) *Response {
	return &Response{Code: c.Code, Message: c.Message}
}

// ErrWithMsg 返回带自定义消息的错误响应。
func ErrWithMsg(c *Code, msg string) *Response {
	return &Response{Code: c.Code, Message: msg}
}

// Respond 发送统一格式的响应到客户端。
func Respond(ctx node.Context, resp *Response) {
	if err := ctx.Response(resp); err != nil {
		log.Errorf("response failed: %v", err)
	}
}

// RespondOK 发送成功响应。
func RespondOK(ctx node.Context) {
	Respond(ctx, OK())
}

// RespondData 发送带数据的成功响应。
func RespondData(ctx node.Context, data any) {
	Respond(ctx, OKWithData(data))
}

// RespondError 发送错误响应。
func RespondError(ctx node.Context, c *Code) {
	Respond(ctx, Err(c))
}

// RespondErrorMsg 发送带自定义消息的错误响应。
func RespondErrorMsg(ctx node.Context, c *Code, msg string) {
	Respond(ctx, ErrWithMsg(c, msg))
}
