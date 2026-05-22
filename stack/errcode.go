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

// ToError 将 Code 转换为 error。
func ToError(c *Code) error {
	if c == nil || c.Code == 0 {
		return nil
	}
	return errors.New(c.Message)
}

// 错误码公式：模块号 × 1000 + 子码，与路由编号统一。
// 系统错误（0–999）复用 HTTP 语义，业务错误按模块分配。

// 系统级错误 (0-999)
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

// Auth 模块错误 (1000-1999)
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

// Player 模块错误 (2000-2999)
var (
	ErrPlayerNotFound = &Code{Code: 2000, Message: "player not found"}
	ErrPlayerBusy     = &Code{Code: 2001, Message: "player is busy"}
	ErrLevelTooLow    = &Code{Code: 2002, Message: "level too low"}
	ErrNameTooLong    = &Code{Code: 2003, Message: "name too long"}
	ErrNotEnoughExp   = &Code{Code: 2004, Message: "not enough experience"}
)

// Chat 模块错误 (3000-3999)
var (
	ErrChatTooFast       = &Code{Code: 3000, Message: "sending too fast"}
	ErrChatBlocked       = &Code{Code: 3001, Message: "user blocked"}
	ErrChatTooLong       = &Code{Code: 3002, Message: "message too long"}
	ErrChatMuted         = &Code{Code: 3003, Message: "you are muted"}
	ErrChatTargetOffline = &Code{Code: 3004, Message: "target offline"}
)

// Match 模块错误 (4000-4999)
var (
	ErrAlreadyInQueue = &Code{Code: 4000, Message: "already in matching queue"}
	ErrNotInQueue     = &Code{Code: 4001, Message: "not in matching queue"}
	ErrMatchTimeout   = &Code{Code: 4002, Message: "match timeout"}
	ErrMatchCanceled  = &Code{Code: 4003, Message: "match canceled"}
	ErrMatchFailed    = &Code{Code: 4004, Message: "match failed"}
)

// Room 模块错误 (5000-5999)
var (
	ErrRoomNotFound  = &Code{Code: 5000, Message: "room not found"}
	ErrRoomFull      = &Code{Code: 5001, Message: "room is full"}
	ErrRoomNotOwner  = &Code{Code: 5002, Message: "not the room owner"}
	ErrRoomAlreadyIn = &Code{Code: 5003, Message: "already in room"}
	ErrRoomNotReady  = &Code{Code: 5004, Message: "not ready"}
	ErrRoomLocked    = &Code{Code: 5005, Message: "room is locked"}
)

// Inventory 模块错误 (6000-6999)
var (
	ErrItemNotFound  = &Code{Code: 6000, Message: "item not found"}
	ErrItemNotEnough = &Code{Code: 6001, Message: "not enough items"}
	ErrBagFull       = &Code{Code: 6002, Message: "bag is full"}
	ErrCannotEquip   = &Code{Code: 6003, Message: "cannot equip item"}
	ErrCannotUse     = &Code{Code: 6004, Message: "cannot use item"}
	ErrSlotLocked    = &Code{Code: 6005, Message: "slot locked"}
)

// Quest 模块错误 (7000-7999)
var (
	ErrQuestNotFound    = &Code{Code: 7000, Message: "quest not found"}
	ErrQuestNotComplete = &Code{Code: 7001, Message: "quest not completed"}
	ErrQuestAlreadyDone = &Code{Code: 7002, Message: "quest already completed"}
	ErrQuestNotAccepted = &Code{Code: 7003, Message: "quest not accepted"}
	ErrQuestCondition   = &Code{Code: 7004, Message: "quest condition not met"}
)

// Combat 模块错误 (8000-8999)
var (
	ErrSkillNotFound = &Code{Code: 8000, Message: "skill not found"}
	ErrSkillCooldown = &Code{Code: 8001, Message: "skill on cooldown"}
	ErrInvalidTarget = &Code{Code: 8002, Message: "invalid target"}
	ErrOutOfRange    = &Code{Code: 8003, Message: "target out of range"}
	ErrNotEnoughMana = &Code{Code: 8004, Message: "not enough mana"}
	ErrCannotCast    = &Code{Code: 8005, Message: "cannot cast skill"}
)

// Guild 模块错误 (9000-9999)
var (
	ErrGuildNotFound         = &Code{Code: 9000, Message: "guild not found"}
	ErrGuildFull             = &Code{Code: 9001, Message: "guild is full"}
	ErrGuildAlreadyMember    = &Code{Code: 9002, Message: "already a guild member"}
	ErrGuildNotMember        = &Code{Code: 9003, Message: "not a guild member"}
	ErrGuildInsufficientRank = &Code{Code: 9004, Message: "insufficient guild rank"}
	ErrGuildNameExists       = &Code{Code: 9005, Message: "guild name already exists"}
	ErrGuildCooldown         = &Code{Code: 9006, Message: "guild cooldown active"}
)

// Shop 模块错误 (11000-11999)
var (
	ErrNotEnoughCurrency = &Code{Code: 11000, Message: "not enough currency"}
	ErrShopItemNotFound  = &Code{Code: 11001, Message: "shop item not found"}
	ErrShopItemSoldOut   = &Code{Code: 11002, Message: "item sold out"}
	ErrShopLevelLocked   = &Code{Code: 11003, Message: "level not high enough"}
	ErrShopLimitReached  = &Code{Code: 11004, Message: "purchase limit reached"}
)

// Leaderboard 模块错误 (12000-12999)
var (
	ErrLeaderboardNotFound = &Code{Code: 12000, Message: "leaderboard not found"}
	ErrRankNotSet          = &Code{Code: 12001, Message: "rank not set"}
)

// Activity 模块错误 (13000-13999)
var (
	ErrActivityNotFound = &Code{Code: 13000, Message: "activity not found"}
	ErrActivityNotOpen  = &Code{Code: 13001, Message: "activity not open"}
	ErrActivityClaimed  = &Code{Code: 13002, Message: "already claimed"}
	ErrActivityExpired  = &Code{Code: 13003, Message: "activity expired"}
)

// Social 模块错误 (14000-14999)
var (
	ErrFriendAlready     = &Code{Code: 14000, Message: "already friends"}
	ErrFriendRequestSent = &Code{Code: 14001, Message: "friend request already sent"}
	ErrFriendSelf        = &Code{Code: 14002, Message: "cannot friend yourself"}
	ErrFriendLimit       = &Code{Code: 14003, Message: "friend list full"}
	ErrBlacklisted       = &Code{Code: 14004, Message: "you have been blocked"}
	ErrNotFriend         = &Code{Code: 14005, Message: "not a friend"}
)
