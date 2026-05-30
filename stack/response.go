package stack

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
	"google.golang.org/protobuf/proto"
)

//==========================
// Proto响应
//==========================

func ProtoResponse(ctx node.Context, pb proto.Message) {
	_ = ctx.Response(pb)
}

//==========================
// JSON响应
//==========================

// Response 统一响应结构。
// 所有客户端消息的回复都使用此结构包裹。
type Response struct {
	Code    int32  `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// JSON 返回成功响应。
func JSON() *Response {
	return &Response{Code: 0, Message: "ok"}
}

// JSONOKWithData 返回带数据的成功响应。
func JSONOKWithData(data any) *Response {
	return &Response{Code: 0, Message: "ok", Data: data}
}

// JSONErr 返回错误响应。
func JSONErr(c *Code) *Response {
	return &Response{Code: c.Code, Message: c.Message}
}

// JSONErrWithMsg 返回带自定义消息的错误响应。
func JSONErrWithMsg(c *Code, msg string) *Response {
	return &Response{Code: c.Code, Message: msg}
}

// JSONRespond 发送统一格式的响应到客户端。
func JSONRespond(ctx node.Context, resp *Response) {
	if err := ctx.Response(resp); err != nil {
		log.Errorf("response failed: %v", err)
	}
}

// JSONRespondOK 发送成功响应。
func JSONRespondOK(ctx node.Context) {
	JSONRespond(ctx, JSON())
}

// JSONRespondData 发送带数据的成功响应。
func JSONRespondData(ctx node.Context, data any) {
	JSONRespond(ctx, JSONOKWithData(data))
}

// JSONRespondError 发送错误响应。
func JSONRespondError(ctx node.Context, c *Code) {
	JSONRespond(ctx, JSONErr(c))
}

// JSONRespondErrorMsg 发送带自定义消息的错误响应。
func JSONRespondErrorMsg(ctx node.Context, c *Code, msg string) {
	JSONRespond(ctx, JSONErrWithMsg(c, msg))
}
