package stack

import "errors"

// Code 定义统一错误码结构。
type Code struct {
	Code    int32
	Message string
}

// Error 实现 error 接口。
func (c *Code) Error() string {
	return c.Message
}

// 系统级错误（0-999），复用 due 框架 codes 对应语义。
var (
	ErrOK               = &Code{Code: 0, Message: "ok"}
	ErrInternalError    = &Code{Code: 500, Message: "internal error"}
	ErrUnauthorized     = &Code{Code: 401, Message: "unauthorized"}
	ErrForbidden        = &Code{Code: 403, Message: "forbidden"}
	ErrNotFound         = &Code{Code: 404, Message: "not found"}
	ErrInvalidParam     = &Code{Code: 400, Message: "invalid parameter"}
	ErrTimeout          = &Code{Code: 408, Message: "request timeout"}
	ErrServerBusy       = &Code{Code: 503, Message: "server busy"}
	ErrDuplicateRequest = &Code{Code: 429, Message: "duplicate request"}
)

// Auth 模块错误 (1000-1099)
var (
	ErrInvalidToken    = &Code{Code: 1000, Message: "invalid token"}
	ErrTokenExpired    = &Code{Code: 1001, Message: "token expired"}
	ErrAccountExists   = &Code{Code: 1002, Message: "account already exists"}
	ErrWrongPassword   = &Code{Code: 1003, Message: "wrong password"}
	ErrAccountBanned   = &Code{Code: 1004, Message: "account banned"}
	ErrLoginElsewhere  = &Code{Code: 1005, Message: "logged in elsewhere"}
	ErrNicknameExists  = &Code{Code: 1006, Message: "nickname already taken"}
	ErrNicknameTooLong = &Code{Code: 1007, Message: "nickname too long"}
)

// Player 模块错误 (1100-1199)
var (
	ErrPlayerNotFound = &Code{Code: 1100, Message: "player not found"}
	ErrPlayerBusy     = &Code{Code: 1101, Message: "player is busy"}
	ErrLevelTooLow    = &Code{Code: 1102, Message: "level too low"}
	ErrNameTooLong    = &Code{Code: 1103, Message: "name too long"}
	ErrNotEnoughExp   = &Code{Code: 1104, Message: "not enough experience"}
)

// Chat 模块错误 (1200-1299)
var (
	ErrChatTooFast       = &Code{Code: 1200, Message: "sending too fast"}
	ErrChatBlocked       = &Code{Code: 1201, Message: "user blocked"}
	ErrChatTooLong       = &Code{Code: 1202, Message: "message too long"}
	ErrChatMuted         = &Code{Code: 1203, Message: "you are muted"}
	ErrChatTargetOffline = &Code{Code: 1204, Message: "target offline"}
)

// Room 模块错误 (1300-1399)
var (
	ErrRoomNotFound  = &Code{Code: 1300, Message: "room not found"}
	ErrRoomFull      = &Code{Code: 1301, Message: "room is full"}
	ErrRoomNotOwner  = &Code{Code: 1302, Message: "not the room owner"}
	ErrRoomAlreadyIn = &Code{Code: 1303, Message: "already in room"}
	ErrRoomNotReady  = &Code{Code: 1304, Message: "not ready"}
	ErrRoomLocked    = &Code{Code: 1305, Message: "room is locked"}
)

// Inventory 模块错误 (1400-1499)
var (
	ErrItemNotFound  = &Code{Code: 1400, Message: "item not found"}
	ErrItemNotEnough = &Code{Code: 1401, Message: "not enough items"}
	ErrBagFull       = &Code{Code: 1402, Message: "bag is full"}
	ErrCannotEquip   = &Code{Code: 1403, Message: "cannot equip item"}
	ErrCannotUse     = &Code{Code: 1404, Message: "cannot use item"}
	ErrSlotLocked    = &Code{Code: 1405, Message: "slot locked"}
)

// Match 模块错误 (1500-1599)
var (
	ErrAlreadyInQueue = &Code{Code: 1500, Message: "already in matching queue"}
	ErrNotInQueue     = &Code{Code: 1501, Message: "not in matching queue"}
	ErrMatchTimeout   = &Code{Code: 1502, Message: "match timeout"}
	ErrMatchCanceled  = &Code{Code: 1503, Message: "match canceled"}
	ErrMatchFailed    = &Code{Code: 1504, Message: "match failed"}
)

// Quest 模块错误 (1600-1699)
var (
	ErrQuestNotFound    = &Code{Code: 1600, Message: "quest not found"}
	ErrQuestNotComplete = &Code{Code: 1601, Message: "quest not completed"}
	ErrQuestAlreadyDone = &Code{Code: 1602, Message: "quest already completed"}
	ErrQuestNotAccepted = &Code{Code: 1603, Message: "quest not accepted"}
	ErrQuestCondition   = &Code{Code: 1604, Message: "quest condition not met"}
)

// Guild 模块错误 (1700-1799)
var (
	ErrGuildNotFound         = &Code{Code: 1700, Message: "guild not found"}
	ErrGuildFull             = &Code{Code: 1701, Message: "guild is full"}
	ErrGuildAlreadyMember    = &Code{Code: 1702, Message: "already a guild member"}
	ErrGuildNotMember        = &Code{Code: 1703, Message: "not a guild member"}
	ErrGuildInsufficientRank = &Code{Code: 1704, Message: "insufficient guild rank"}
	ErrGuildNameExists       = &Code{Code: 1705, Message: "guild name already exists"}
	ErrGuildCooldown         = &Code{Code: 1706, Message: "guild cooldown active"}
)

// Shop 模块错误 (1800-1899)
var (
	ErrNotEnoughCurrency = &Code{Code: 1800, Message: "not enough currency"}
	ErrShopItemNotFound  = &Code{Code: 1801, Message: "shop item not found"}
	ErrShopItemSoldOut   = &Code{Code: 1802, Message: "item sold out"}
	ErrShopLevelLocked   = &Code{Code: 1803, Message: "level not high enough"}
	ErrShopLimitReached  = &Code{Code: 1804, Message: "purchase limit reached"}
)

// Social 模块错误 (1900-1999)
var (
	ErrFriendAlready     = &Code{Code: 1900, Message: "already friends"}
	ErrFriendRequestSent = &Code{Code: 1901, Message: "friend request already sent"}
	ErrFriendSelf        = &Code{Code: 1902, Message: "cannot friend yourself"}
	ErrFriendLimit       = &Code{Code: 1903, Message: "friend list full"}
	ErrBlacklisted       = &Code{Code: 1904, Message: "you have been blocked"}
	ErrNotFriend         = &Code{Code: 1905, Message: "not a friend"}
)

// Activity 模块错误 (2000-2099)
var (
	ErrActivityNotFound = &Code{Code: 2000, Message: "activity not found"}
	ErrActivityNotOpen  = &Code{Code: 2001, Message: "activity not open"}
	ErrActivityClaimed  = &Code{Code: 2002, Message: "already claimed"}
	ErrActivityExpired  = &Code{Code: 2003, Message: "activity expired"}
)

// Combat 模块错误 (2100-2199)
var (
	ErrSkillNotFound = &Code{Code: 2100, Message: "skill not found"}
	ErrSkillCooldown = &Code{Code: 2101, Message: "skill on cooldown"}
	ErrInvalidTarget = &Code{Code: 2102, Message: "invalid target"}
	ErrOutOfRange    = &Code{Code: 2103, Message: "target out of range"}
	ErrNotEnoughMana = &Code{Code: 2104, Message: "not enough mana"}
	ErrCannotCast    = &Code{Code: 2105, Message: "cannot cast skill"}
)

// Leaderboard 模块错误 (2200-2299)
var (
	ErrLeaderboardNotFound = &Code{Code: 2200, Message: "leaderboard not found"}
	ErrRankNotSet          = &Code{Code: 2201, Message: "rank not set"}
)

// ToError 将 Code 转换为 error。
func ToError(c *Code) error {
	if c == nil || c.Code == 0 {
		return nil
	}
	return errors.New(c.Message)
}
