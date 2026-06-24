package middleware

import (
	"strings"

	"github.com/dobyte/due/component/http/v2"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/utils/xconv"
	"github.com/dobyte/jwt"

	jwtComponent "github.com/skeletongo/game-stack/internal/component/jwt"
)

const (
	CtxToken  = "jwt_token"              // 上下文中的 JWT token
	CtxUID    = "jwt_uid"                // 上下文中的用户id
	CtxGameID = jwtComponent.ClaimGameID // 上下文中的游戏id
)

// NewAuth 创建鉴权中间件。
func NewAuth(proxy *http.Proxy, jwt *jwt.JWT) *Auth {
	return &Auth{proxy: proxy, jwt: jwt}
}

// Auth 封装 HTTP 鉴权依赖。
type Auth struct {
	proxy *http.Proxy // HTTP 代理
	jwt   *jwt.JWT    // JWT 组件
}

// JwtAuth 校验 Authorization 头并写入用户上下文。
func (a *Auth) JwtAuth(ctx http.Context) error {
	reqHeader := ctx.GetReqHeaders()
	if authHeader, ok := reqHeader["Authorization"]; ok {
		if len(authHeader) == 0 {
			log.Warnf("need Authorization: %#v", reqHeader)
			return ctx.SendStatus(http.StatusUnauthorized)
		}
		tokenString := strings.Replace(authHeader[0], "Bearer ", "", 1)

		payload, err := a.jwt.ParseToken(tokenString)
		if err != nil {
			log.Warnf("")
			return ctx.SendStatus(http.StatusUnauthorized)
		}
		uid := xconv.Int64(payload.Subject())
		gameID := xconv.Int64(payload[CtxGameID])

		ctx.RequestCtx().SetUserValue(CtxToken, tokenString)
		ctx.RequestCtx().SetUserValue(CtxUID, uid) // 外部玩家id
		ctx.RequestCtx().SetUserValue(CtxGameID, gameID)
		log.Debugf("JwtAuth gameId:%d ,Uid:%d, Bearer %s", gameID, uid, tokenString)
	} else {
		return ctx.SendStatus(http.StatusUnauthorized)
	}

	return ctx.Next()
}

// GetUID 从 HTTP 上下文获取用户id。
func GetUID(ctx http.Context) int64 {
	return xconv.Int64(ctx.RequestCtx().UserValue(CtxUID))
}

// GetToken 从 HTTP 上下文获取登录令牌。
func GetToken(ctx http.Context) string {
	return xconv.String(ctx.RequestCtx().UserValue(CtxToken))
}

// GetGameID 从 HTTP 上下文获取游戏id。
func GetGameID(ctx http.Context) int64 {
	return xconv.Int64(ctx.RequestCtx().UserValue(CtxGameID))
}
