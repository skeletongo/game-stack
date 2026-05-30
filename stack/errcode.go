package stack

import (
	"errors"

	"github.com/skeletongo/game-stack/proto/auth"
	"github.com/skeletongo/game-stack/proto/common"
	"github.com/skeletongo/game-stack/proto/player"
)

// Code 定义统一错误码结构。
type Code struct {
	Code    int32
	Message string
}

// Error 实现 error 接口。
func (c *Code) Error() string {
	return c.Message
}

// ToError 将 Code 转换为 error。
func ToError(c *Code) error {
	if c == nil || c.Code == 0 {
		return nil
	}
	return errors.New(c.Message)
}

// ErrCode 从 error 中提取错误码，若 err 是 *Code 则返回其 Code，否则返回内部错误。
func ErrCode(err error) int32 {
	var c *Code
	if errors.As(err, &c) {
		return c.Code
	}
	return int32(common.SysError_INTERNAL_ERROR)
}

// 错误码公式：模块号 × 1000 + 子码，与路由编号统一。
// 值由各模块 proto 文件的 Error 枚举定义，客户端与服务端共享。

// 系统级错误 (0-999)
var (
	CodeOK              = int32(common.SysError_OK)
	CodeErr             = int32(common.SysError_INTERNAL_ERROR)
	ErrOK               = &Code{Code: int32(common.SysError_OK), Message: "ok"}
	ErrInvalidParam     = &Code{Code: int32(common.SysError_INVALID_PARAM), Message: "invalid parameter"}
	ErrUnauthorized     = &Code{Code: int32(common.SysError_UNAUTHORIZED), Message: "unauthorized"}
	ErrForbidden        = &Code{Code: int32(common.SysError_FORBIDDEN), Message: "forbidden"}
	ErrNotFound         = &Code{Code: int32(common.SysError_NOT_FOUND), Message: "not found"}
	ErrTimeout          = &Code{Code: int32(common.SysError_TIMEOUT), Message: "request timeout"}
	ErrInternalError    = &Code{Code: int32(common.SysError_INTERNAL_ERROR), Message: "internal error"}
	ErrDuplicateRequest = &Code{Code: int32(common.SysError_DUPLICATE_REQUEST), Message: "duplicate request"}
	ErrServerBusy       = &Code{Code: int32(common.SysError_SERVER_BUSY), Message: "server busy"}
)

// Auth 模块错误 (1000-1999)
var (
	ErrInvalidToken    = &Code{Code: int32(auth.AuthError_INVALID_TOKEN), Message: "invalid token"}
	ErrTokenExpired    = &Code{Code: int32(auth.AuthError_TOKEN_EXPIRED), Message: "token expired"}
	ErrAccountExists   = &Code{Code: int32(auth.AuthError_ACCOUNT_EXISTS), Message: "account already exists"}
	ErrWrongPassword   = &Code{Code: int32(auth.AuthError_WRONG_PASSWORD), Message: "wrong password"}
	ErrAccountBanned   = &Code{Code: int32(auth.AuthError_ACCOUNT_BANNED), Message: "account banned"}
	ErrLoginElsewhere  = &Code{Code: int32(auth.AuthError_LOGIN_ELSEWHERE), Message: "logged in elsewhere"}
	ErrNicknameExists  = &Code{Code: int32(auth.AuthError_NICKNAME_EXISTS), Message: "nickname already taken"}
	ErrNicknameTooLong = &Code{Code: int32(auth.AuthError_NICKNAME_TOO_LONG), Message: "nickname too long"}
	ErrInvalidUsername = &Code{Code: int32(auth.AuthError_INVALID_USERNAME), Message: "invalid username"}
	ErrInvalidNickname = &Code{Code: int32(auth.AuthError_INVALID_NICKNAME), Message: "invalid nickname"}
)

// Player 模块错误 (2000-2999)
var (
	ErrPlayerNotFound      = &Code{Code: int32(player.PlayerError_PLAYER_NOT_FOUND), Message: "player not found"}
	ErrInvalidLevel        = &Code{Code: int32(player.PlayerError_INVALID_LEVEL), Message: "invalid level"}
	ErrInsufficientGold    = &Code{Code: int32(player.PlayerError_INSUFFICIENT_GOLD), Message: "insufficient gold"}
	ErrNameTooLong         = &Code{Code: int32(player.PlayerError_NAME_TOO_LONG), Message: "name too long"}
	ErrInsufficientDiamond = &Code{Code: int32(player.PlayerError_INSUFFICIENT_DIAMOND), Message: "insufficient diamond"}
	ErrInvalidExp          = &Code{Code: int32(player.PlayerError_INVALID_EXP), Message: "invalid exp"}
)
