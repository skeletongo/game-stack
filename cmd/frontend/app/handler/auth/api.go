package auth

import (
	"github.com/dobyte/due/component/http/v2"

	"github.com/skeletongo/game-stack/cmd/frontend/app/context"
	"github.com/skeletongo/game-stack/cmd/frontend/app/handler/auth/types"
	"github.com/skeletongo/game-stack/cmd/frontend/app/middleware"
	authgrpc "github.com/skeletongo/game-stack/module/auth/rpc/grpc"
	"github.com/skeletongo/game-stack/stack"
)

// api 提供 auth HTTP 接口。
type api struct {
	*context.ServiceContext // 服务上下文
}

// New 创建 auth HTTP 接口处理器。
func New() *api {
	return &api{
		ServiceContext: context.GetSvc(),
	}
}

// Init 注册 auth HTTP 路由。
func (a *api) Init() {
	router := a.ProxyHttp.Router()
	// 注册
	router.Post("/register", a.register)
	// 登录
	router.Post("/login", a.login)

	// user 接口
	user := router.Group("/user", a.Auth.JwtAuth)
	// 登出
	user.Post("/logout", a.logout)
	// 刷新token
	user.Post("/token_refresh", a.tokenRefresh)
}

// Register
// @Summary 注册
// @Description
// @Tags AUTH
// @Accept json
// @Produce json
// @Param data body	types.RegisterReq true "请求参数"
// @Success 200 {object} types.RegisterResp
// @Router /register [post]
func (a *api) register(ctx http.Context) error {
	req := &types.RegisterReq{}
	if err := ctx.Bind().Body(req); err != nil {
		return ctx.JSON(types.RegisterResp{Code: stack.ErrInvalidParam.Code, Message: err.Error()})
	}

	authClient, err := a.AuthClient()
	if err != nil {
		return ctx.JSON(types.RegisterResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	resp, err := authClient.Register(ctx.Context(), &authgrpc.RegisterReq{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		return ctx.JSON(types.RegisterResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	return ctx.JSON(types.RegisterResp{Code: resp.GetCode(), Message: resp.GetMessage()})
}

// Login
// @Summary 登录
// @Description
// @Tags AUTH
// @Accept json
// @Produce json
// @Param data body types.LoginReq true "请求参数"
// @Success 200 {object} types.LoginResp
// @Router /login [post]
func (a *api) login(ctx http.Context) error {
	req := &types.LoginReq{}
	if err := ctx.Bind().Body(req); err != nil {
		return ctx.JSON(types.LoginResp{Code: stack.ErrInvalidParam.Code, Message: err.Error()})
	}

	authClient, err := a.AuthClient()
	if err != nil {
		return ctx.JSON(types.LoginResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	resp, err := authClient.Login(ctx.Context(), &authgrpc.LoginReq{
		Username: req.Username,
		Password: req.Password,
		GameId:   req.GameID,
	})
	if err != nil {
		return ctx.JSON(types.LoginResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	return ctx.JSON(types.LoginResp{
		Code:      resp.GetCode(),
		Message:   resp.GetMessage(),
		Token:     resp.GetToken(),
		ExpiresAt: resp.GetExpiresAt(),
		PlayerID:  resp.GetPlayerId(),
		UnixNano:  resp.GetUnixNano(),
	})
}

// Logout
// @Summary 登出
// @Description
// @Tags AUTH
// @Accept json
// @Produce json
// @Success 200 {object} types.LogoutResp
// @Router /user/logout [post]
func (a *api) logout(ctx http.Context) error {
	authClient, err := a.AuthClient()
	if err != nil {
		return ctx.JSON(types.LogoutResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	resp, err := authClient.Logout(ctx.Context(), &authgrpc.LogoutReq{UserId: middleware.GetUID(ctx)})
	if err != nil {
		return ctx.JSON(types.LogoutResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	return ctx.JSON(types.LogoutResp{Code: resp.GetCode(), Message: resp.GetMessage()})
}

// TokenRefresh
// @Summary 刷新token
// @Description
// @Tags AUTH
// @Accept json
// @Produce json
// @Success 200 {object} types.TokenRefreshResp
// @Router /user/token_refresh [post]
func (a *api) tokenRefresh(ctx http.Context) error {
	authClient, err := a.AuthClient()
	if err != nil {
		return ctx.JSON(types.TokenRefreshResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	resp, err := authClient.RefreshToken(ctx.Context(), &authgrpc.RefreshTokenReq{
		UserId: middleware.GetUID(ctx),
		Token:  middleware.GetToken(ctx),
	})
	if err != nil {
		return ctx.JSON(types.TokenRefreshResp{Code: stack.ErrInternalError.Code, Message: err.Error()})
	}
	return ctx.JSON(types.TokenRefreshResp{
		Code:      resp.GetCode(),
		Message:   resp.GetMessage(),
		Token:     resp.GetToken(),
		ExpiresAt: resp.GetExpiresAt(),
	})
}
