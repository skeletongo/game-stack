package server

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/auth/internal/application"
	pb "github.com/skeletongo/game-stack/module/auth/rpc/grpc"
	"github.com/skeletongo/game-stack/stack"
)

// Register 注册 auth RPC 服务提供者。
func Register(name string, proxy *node.Proxy, cmdBus *ddd.CommandBus) {
	proxy.AddServiceProvider(name, &pb.Auth_ServiceDesc, &server{cmdBus: cmdBus})
}

// server 实现短连接 auth RPC 服务。
type server struct {
	pb.UnimplementedAuthServer
	cmdBus *ddd.CommandBus // 命令总线
}

// Register 处理短连接注册请求。
func (s *server) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return &pb.RegisterResp{Code: stack.ErrInvalidParam.Code, Message: stack.ErrInvalidParam.Message}, nil
	}
	nickname := req.GetNickname()
	if nickname == "" {
		nickname = req.GetUsername()
	}

	result, err := ddd.Dispatch[*application.RegisterResult](ctx, s.cmdBus, application.RegisterCmd{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Nickname: nickname,
	})
	if err != nil {
		return &pb.RegisterResp{Code: stack.ErrCode(err), Message: err.Error()}, nil
	}
	if result.UserID <= 0 {
		return &pb.RegisterResp{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message}, nil
	}
	return &pb.RegisterResp{Code: stack.CodeOK}, nil
}

// Login 处理短连接登录请求，只签发 token，不绑定长连接会话。
func (s *server) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return &pb.LoginResp{Code: stack.ErrInvalidParam.Code, Message: "username and password required"}, nil
	}

	result, err := ddd.Dispatch[*application.LoginResult](ctx, s.cmdBus, application.LoginCmd{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		GameID:   req.GetGameId(),
	})
	if err != nil {
		return &pb.LoginResp{Code: stack.ErrCode(err), Message: err.Error()}, nil
	}

	return &pb.LoginResp{
		Code:      stack.CodeOK,
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		PlayerId:  result.PlayerID,
		UnixNano:  time.Now().UnixNano(),
	}, nil
}

// Logout 处理短连接登出请求。
func (s *server) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	if req.GetUserId() == 0 {
		return &pb.LogoutResp{Code: stack.ErrUnauthorized.Code, Message: stack.ErrUnauthorized.Message}, nil
	}
	if _, err := s.cmdBus.Dispatch(ctx, application.LogoutCmd{UserID: req.GetUserId()}); err != nil {
		return &pb.LogoutResp{Code: stack.ErrCode(err), Message: err.Error()}, nil
	}
	return &pb.LogoutResp{Code: stack.CodeOK}, nil
}

// RefreshToken 处理短连接 token 刷新请求。
func (s *server) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenResp, error) {
	if req.GetUserId() == 0 || req.GetToken() == "" {
		return &pb.RefreshTokenResp{Code: stack.ErrInvalidParam.Code, Message: stack.ErrInvalidParam.Message}, nil
	}
	result, err := ddd.Dispatch[*application.RefreshTokenResult](ctx, s.cmdBus, application.RefreshTokenCmd{
		UserID: req.GetUserId(),
		Token:  req.GetToken(),
	})
	if err != nil {
		return &pb.RefreshTokenResp{Code: stack.ErrCode(err), Message: err.Error()}, nil
	}
	return &pb.RefreshTokenResp{Code: stack.CodeOK, Token: result.Token, ExpiresAt: result.ExpiresAt}, nil
}
