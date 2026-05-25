package stack

import "github.com/dobyte/due/v2/cluster/node"

// StatefulAuthorizedRoute 组合有状态 + 授权路由选项。
// 有状态路由确保同一玩家的所有请求被路由到其绑定的固定节点，
// 授权路由确保请求必须携带合法的 UID。
var StatefulAuthorizedRoute = node.RouteOptions{
	Stateful:   true,
	Authorized: true,
}

// 路由编号公式：模块号 × 1000 + 子协议号
// 模块号 0（0–999）为系统内部路由预留。
//
// 模块号分配：
//
//	 0 — 系统内部
//	 1 — auth         2 — player       3 — chat         4 — match
//	 5 — room         6 — inventory    7 — quest        8 — combat
//	 9 — guild       10 — mail        11 — shop        12 — leaderboard
//	13 — activity    14 — social
const (
	// Auth 模块 (1000-1999)
	RouteAuthLogin        int32 = 1001
	RouteAuthRegister     int32 = 1002
	RouteAuthLogout       int32 = 1003
	RouteAuthTokenRefresh int32 = 1004
	RouteAuthKick         int32 = 1005

	// Player 模块 (2000-2999)
	RoutePlayerGetInfo   int32 = 2001
	RoutePlayerUpdate    int32 = 2002
	RoutePlayerDelete    int32 = 2003
	RoutePlayerSearch    int32 = 2004
	RoutePlayerSetAvatar int32 = 2005
	RoutePlayerLevelUp   int32 = 2006

	// Chat 模块 (3000-3999)
	RouteChatSend    int32 = 3001
	RouteChatHistory int32 = 3002
	RouteChatPrivate int32 = 3003
	RouteChatWorld   int32 = 3004
	RouteChatGuild   int32 = 3005

	// Match 模块 (4000-4999)
	RouteMatchJoin   int32 = 4001
	RouteMatchLeave  int32 = 4002
	RouteMatchStatus int32 = 4003
	RouteMatchCancel int32 = 4004

	// Room 模块 (5000-5999)
	RouteRoomCreate int32 = 5001
	RouteRoomJoin   int32 = 5002
	RouteRoomLeave  int32 = 5003
	RouteRoomList   int32 = 5004
	RouteRoomInfo   int32 = 5005
	RouteRoomReady  int32 = 5006
	RouteRoomKick   int32 = 5007

	// Inventory 模块 (6000-6999)
	RouteInvList    int32 = 6001
	RouteInvUse     int32 = 6002
	RouteInvEquip   int32 = 6003
	RouteInvUnequip int32 = 6004
	RouteInvDrop    int32 = 6005
	RouteInvSell    int32 = 6006

	// Quest 模块 (7000-7999)
	RouteQuestList     int32 = 7001
	RouteQuestAccept   int32 = 7002
	RouteQuestSubmit   int32 = 7003
	RouteQuestAbandon  int32 = 7004
	RouteQuestProgress int32 = 7005

	// Combat 模块 (8000-8999)
	RouteCombatSkillCast int32 = 8001
	RouteCombatMove      int32 = 8002
	RouteCombatStateSync int32 = 8003
	RouteCombatTarget    int32 = 8004

	// Guild 模块 (9000-9999)
	RouteGuildCreate  int32 = 9001
	RouteGuildJoin    int32 = 9002
	RouteGuildLeave   int32 = 9003
	RouteGuildKick    int32 = 9004
	RouteGuildList    int32 = 9005
	RouteGuildInfo    int32 = 9006
	RouteGuildDonate  int32 = 9007
	RouteGuildUpgrade int32 = 9008

	// Mail 模块 (10000-10999)
	RouteMailList          int32 = 10001
	RouteMailRead          int32 = 10002
	RouteMailReceiveAttach int32 = 10003
	RouteMailDelete        int32 = 10004
	RouteMailSend          int32 = 10005

	// Shop 模块 (11000-11999)
	RouteShopList     int32 = 11001
	RouteShopBuy      int32 = 11002
	RouteShopBuyBatch int32 = 11003
	RouteShopRefresh  int32 = 11004

	// Leaderboard 模块 (12000-12999)
	RouteLeaderboardGet  int32 = 12001
	RouteLeaderboardRank int32 = 12002
	RouteLeaderboardTop  int32 = 12003

	// Activity 模块 (13000-13999)
	RouteActivityList     int32 = 13001
	RouteActivityClaim    int32 = 13002
	RouteActivityInfo     int32 = 13003
	RouteActivityProgress int32 = 13004

	// Social 模块 (14000-14999)
	RouteSocialFriendList   int32 = 14001
	RouteSocialFriendAdd    int32 = 14002
	RouteSocialFriendRemove int32 = 14003
	RouteSocialBlacklist    int32 = 14004
	RouteSocialBlock        int32 = 14005
	RouteSocialUnblock      int32 = 14006
)
