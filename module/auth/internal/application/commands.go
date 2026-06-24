package application

// 命令名称常量。
const (
	CmdRegister     = "auth.register"
	CmdLogin        = "auth.login"
	CmdTokenLogin   = "auth.token_login"
	CmdLogout       = "auth.logout"
	CmdRefreshToken = "auth.refresh_token"
)

// RegisterCmd 注册命令。
type RegisterCmd struct {
	Username string // 账号名
	Password string // 密码
	Nickname string // 昵称
}

func (c RegisterCmd) CommandName() string { return CmdRegister }

// LoginCmd 登录命令。
type LoginCmd struct {
	Username string // 账号名
	Password string // 密码
	GameID   int64  // 游戏id
}

func (c LoginCmd) CommandName() string { return CmdLogin }

// TokenLoginCmd 使用 token 登录长连接。
type TokenLoginCmd struct {
	Token string // 登录令牌
}

func (c TokenLoginCmd) CommandName() string { return CmdTokenLogin }

// LogoutCmd 登出命令。
type LogoutCmd struct {
	UserID int64 // 用户id
}

func (c LogoutCmd) CommandName() string { return CmdLogout }

// RefreshTokenCmd 刷新令牌命令。
type RefreshTokenCmd struct {
	UserID int64  // 用户id
	Token  string // 原登录令牌
}

func (c RefreshTokenCmd) CommandName() string { return CmdRefreshToken }
