package application

// 命令名称常量。
const (
	CmdCreatePlayer  = "player.create"
	CmdGetPlayer     = "player.get_info"
	CmdSetAvatar     = "player.set_avatar"
	CmdDeletePlayer  = "player.delete"
	CmdAddExp        = "player.add_exp"
	CmdAddGold       = "player.add_gold"
	CmdDeductGold    = "player.deduct_gold"
	CmdAddDiamond    = "player.add_diamond"
	CmdDeductDiamond = "player.deduct_diamond"
)

// CreatePlayerCmd 创建玩家命令（注册账号时由 auth 模块调用）。
type CreatePlayerCmd struct {
	PlayerID int64
	Nickname string
}

func (c CreatePlayerCmd) CommandName() string { return CmdCreatePlayer }

// GetPlayerCmd 查询玩家信息命令（只读，不走 Actor）。
type GetPlayerCmd struct {
	PlayerID int64
	TargetID int64 // 为 0 时查询自身
}

func (c GetPlayerCmd) CommandName() string { return CmdGetPlayer }

// SetAvatarCmd 设置头像命令。
type SetAvatarCmd struct {
	PlayerID int64
	Avatar   string
}

func (c SetAvatarCmd) CommandName() string { return CmdSetAvatar }

// DeletePlayerCmd 删除玩家命令（内部使用）。
type DeletePlayerCmd struct{ PlayerID int64 }

func (c DeletePlayerCmd) CommandName() string { return CmdDeletePlayer }

// AddExpCmd 增加经验值命令。
type AddExpCmd struct {
	PlayerID int64
	Amount   int64
}

func (c AddExpCmd) CommandName() string { return CmdAddExp }

// AddGoldCmd 增加金币命令。
type AddGoldCmd struct {
	PlayerID int64
	Amount   int32
}

func (c AddGoldCmd) CommandName() string { return CmdAddGold }

// DeductGoldCmd 扣除金币命令。
type DeductGoldCmd struct {
	PlayerID int64
	Amount   int32
}

func (c DeductGoldCmd) CommandName() string { return CmdDeductGold }

// AddDiamondCmd 增加钻石命令。
type AddDiamondCmd struct {
	PlayerID int64
	Amount   int32
}

func (c AddDiamondCmd) CommandName() string { return CmdAddDiamond }

// DeductDiamondCmd 扣除钻石命令。
type DeductDiamondCmd struct {
	PlayerID int64
	Amount   int32
}

func (c DeductDiamondCmd) CommandName() string { return CmdDeductDiamond }
