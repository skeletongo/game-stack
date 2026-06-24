package middleware

import (
	"strings"

	"github.com/dobyte/due/component/http/v2"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/mode"
)

// Log 在调试模式下输出 HTTP 请求和响应日志。
func Log(ctx http.Context) error {
	if mode.IsDebugMode() {
		defer func() {
			path := string(ctx.Request().URI().Path())
			headers := ctx.GetReqHeaders()
			reqBody := string(ctx.Request().Body())
			respType := ctx.GetRespHeader("Content-Type")
			if strings.HasPrefix(respType, "text/event-stream") {
				log.Debugf("path:%s reqHeaders:%v reqBody:%s respType:%s respBody:<stream>", path, headers, reqBody, respType)
				return
			}

			status := ctx.Response().StatusCode()
			respBody := string(ctx.Response().Body())
			log.Debugf("path:%s status:%d reqHeaders:%v reqBody:%s respType:%s respBody:%s", path, status, headers, reqBody, respType, respBody)
		}()
	}
	return ctx.Next()
}
