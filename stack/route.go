package stack

import "github.com/dobyte/due/v2/cluster/node"

// StatefulAuthorizedRoute 组合有状态 + 授权路由选项。
// 有状态路由确保同一玩家的所有请求被路由到其绑定的固定节点，
// 授权路由确保请求必须携带合法的 UID。
var StatefulAuthorizedRoute = node.RouteOptions{
	Stateful:   true,
	Authorized: true,
}

// 统一路由常量。每模块预留100个号段，避免冲突。
const (
	// Auth 模块 (1-99)
	RouteAuthLogin        int32 = 1
	RouteAuthRegister     int32 = 2
	RouteAuthLogout       int32 = 3
	RouteAuthTokenRefresh int32 = 4
	RouteAuthKick         int32 = 5

	// Player 模块 (101-199)
	RoutePlayerGetInfo   int32 = 101
	RoutePlayerUpdate    int32 = 102
	RoutePlayerDelete    int32 = 103
	RoutePlayerSearch    int32 = 104
	RoutePlayerSetAvatar int32 = 105
	RoutePlayerLevelUp   int32 = 106

	// Chat 模块 (201-299)
	RouteChatSend    int32 = 201
	RouteChatHistory int32 = 202
	RouteChatPrivate int32 = 203
	RouteChatWorld   int32 = 204
	RouteChatGuild   int32 = 205

	// Match 模块 (301-399)
	RouteMatchJoin   int32 = 301
	RouteMatchLeave  int32 = 302
	RouteMatchStatus int32 = 303
	RouteMatchCancel int32 = 304

	// Room 模块 (401-499)
	RouteRoomCreate int32 = 401
	RouteRoomJoin   int32 = 402
	RouteRoomLeave  int32 = 403
	RouteRoomList   int32 = 404
	RouteRoomInfo   int32 = 405
	RouteRoomReady  int32 = 406
	RouteRoomKick   int32 = 407

	// Inventory 模块 (501-599)
	RouteInvList    int32 = 501
	RouteInvUse     int32 = 502
	RouteInvEquip   int32 = 503
	RouteInvUnequip int32 = 504
	RouteInvDrop    int32 = 505
	RouteInvSell    int32 = 506

	// Quest 模块 (601-699)
	RouteQuestList     int32 = 601
	RouteQuestAccept   int32 = 602
	RouteQuestSubmit   int32 = 603
	RouteQuestAbandon  int32 = 604
	RouteQuestProgress int32 = 605

	// Combat 模块 (701-799)
	RouteCombatSkillCast int32 = 701
	RouteCombatMove      int32 = 702
	RouteCombatStateSync int32 = 703
	RouteCombatTarget    int32 = 704

	// Guild 模块 (801-899)
	RouteGuildCreate  int32 = 801
	RouteGuildJoin    int32 = 802
	RouteGuildLeave   int32 = 803
	RouteGuildKick    int32 = 804
	RouteGuildList    int32 = 805
	RouteGuildInfo    int32 = 806
	RouteGuildDonate  int32 = 807
	RouteGuildUpgrade int32 = 808

	// Mail 模块 (901-999)
	RouteMailList          int32 = 901
	RouteMailRead          int32 = 902
	RouteMailReceiveAttach int32 = 903
	RouteMailDelete        int32 = 904
	RouteMailSend          int32 = 905

	// Shop 模块 (1001-1099)
	RouteShopList     int32 = 1001
	RouteShopBuy      int32 = 1002
	RouteShopBuyBatch int32 = 1003
	RouteShopRefresh  int32 = 1004

	// Leaderboard 模块 (1101-1199)
	RouteLeaderboardGet  int32 = 1101
	RouteLeaderboardRank int32 = 1102
	RouteLeaderboardTop  int32 = 1103

	// Activity 模块 (1201-1299)
	RouteActivityList     int32 = 1201
	RouteActivityClaim    int32 = 1202
	RouteActivityInfo     int32 = 1203
	RouteActivityProgress int32 = 1204

	// Social 模块 (1301-1399)
	RouteSocialFriendList   int32 = 1301
	RouteSocialFriendAdd    int32 = 1302
	RouteSocialFriendRemove int32 = 1303
	RouteSocialBlacklist    int32 = 1304
	RouteSocialBlock        int32 = 1305
	RouteSocialUnblock      int32 = 1306
)
