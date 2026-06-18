package application

// 命令名称常量。
const (
	CmdRegister     = "auth.register"
	CmdLogin        = "auth.login"
	CmdLogout       = "auth.logout"
	CmdRefreshToken = "auth.refresh_token"
	CmdMarkOnline   = "auth.mark_online"
)

// RegisterCmd 注册命令。
type RegisterCmd struct {
	Username string
	Password string
	Nickname string
}

func (c RegisterCmd) CommandName() string { return CmdRegister }

// LoginCmd 登录命令。
type LoginCmd struct {
	Username string
	Password string
}

func (c LoginCmd) CommandName() string { return CmdLogin }

type MarkOnlineCmd struct {
	UserID int64
	Token  string
	GID    string
}

func (c MarkOnlineCmd) CommandName() string { return CmdMarkOnline }

// LogoutCmd 登出命令。
type LogoutCmd struct{ UserID int64 }

func (c LogoutCmd) CommandName() string { return CmdLogout }

// RefreshTokenCmd 刷新令牌命令。
type RefreshTokenCmd struct {
	UserID int64
	Token  string
}

func (c RefreshTokenCmd) CommandName() string { return CmdRefreshToken }
