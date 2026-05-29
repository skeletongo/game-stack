package stack

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/skeletongo/game-stack/proto/auth"
	"github.com/skeletongo/game-stack/proto/player"
)

// StatefulAuthorizedRoute 组合有状态 + 授权路由选项。
// 有状态路由确保同一玩家的所有请求被路由到其绑定的固定节点，
// 授权路由确保请求必须携带合法的 UID。
var StatefulAuthorizedRoute = node.RouteOptions{
	Stateful:   true,
	Authorized: true,
}

// 路由编号常量，值由各模块 proto 文件的 Route 枚举定义，客户端与服务端共享。
//
// 公式：模块号 × 1000 + 子协议号
// 模块号 0（0–999）为系统内部路由预留。
const (
	// Auth 模块 (1000-1999)
	RouteAuthLogin        int32 = int32(auth.AuthRoute_LOGIN)
	RouteAuthRegister     int32 = int32(auth.AuthRoute_REGISTER)
	RouteAuthLogout       int32 = int32(auth.AuthRoute_LOGOUT)
	RouteAuthTokenRefresh int32 = int32(auth.AuthRoute_TOKEN_REFRESH)
	RouteAuthKick         int32 = int32(auth.AuthRoute_KICK)

	// Player 模块 (2000-2999)
	RoutePlayerGetInfo   int32 = int32(player.PlayerRoute_GET_INFO)
	RoutePlayerSetAvatar int32 = int32(player.PlayerRoute_SET_AVATAR)
	RoutePlayerLevelUp   int32 = int32(player.PlayerRoute_LEVEL_UP)
)
