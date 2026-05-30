package tool

// RouteInfo 表示 stack/route.go 中解析出的一个路由常量。
type RouteInfo struct {
	Name       string // 常量名，如 "RouteAuthLogin"
	Number     int32  // 路由号，如 1001
	ModuleName string // 所属模块名，如 "auth"
	Action     string // 操作名（去掉 Route 和模块名前缀后剩余的部分），如 "Login"
	Line       int    // 在 route.go 中的行号 (1-based)
}

// ModuleSection 表示 route.go 中一个模块的路由段。
// 包含段头部注释（如 "// Auth 模块 (1000-1999)"）和其下所有路由。
type ModuleSection struct {
	ModuleName   string      // 模块名，如 "auth"
	ModuleNumber int         // 模块号，如 1
	BaseNumber   int32       // 该段起始路由号，如 1000
	HeaderLine   int         // 段头部注释行号 (1-based)
	Routes       []RouteInfo // 该段内所有路由（按行号排序）
	EndLine      int         // 最后一行路由常量或头部注释的行号（用于插入点定位）
}

// ErrRange 表示 stack/errcode.go 中一个模块的错误码 var() 块。
type ErrRange struct {
	Module string    // 模块名
	Base   int32     // 最小错误码（向下取整到 100 的倍数）
	Errs   []ErrInfo // 该块内所有错误码
	EndLine int      // var() 块的结束行号 (1-based)
}

// ErrInfo 表示一个错误码常量。
type ErrInfo struct {
	Name    string // 常量名，如 "ErrInvalidToken"
	Code    int32  // 错误码
	Message string // 错误消息
}

// ModuleInfo 聚合一个游戏模块的所有信息。
type ModuleInfo struct {
	Name     string      // 模块名，如 "auth"
	Number   int         // 模块号，如 1
	Routes   []RouteInfo // 该模块已有路由
	ErrRange *ErrRange   // 该模块已有错误码块 (nil 表示无)
	DirExist bool        // module/<name>/ 目录是否存在
}

// RouteType 路由注册类型。
type RouteType int

const (
	RouteStandard RouteType = iota // proxy.AddRouteHandler(route, handler, stack.StatefulAuthorizedRoute)
	RouteNoAuth                    // proxy.AddRouteHandler(route, handler)
	RouteActor                     // proxy.AddRouteHandler(route, actor.RouteToActor(...), stack.StatefulAuthorizedRoute)
)

// ModuleGenData 生成模块代码时的模板数据。
type ModuleGenData struct {
	Name      string // 包名，如 "example"
	NameTitle string // 导出名，如 "Example"
	Number    int    // 模块号
	ErrBase   int32  // 错误码基数
}
