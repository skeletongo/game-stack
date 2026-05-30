// Package debug 提供开发调试用的 HTTP 服务。
//
// 端点：
//   - GET  /debug/modules           列出所有模块
//   - GET  /debug/module/:name       列出模块能力
//   - POST /debug/query              查询聚合快照
//   - POST /debug/command            执行命令
//   - POST /debug/patch              直接修改聚合字段
//
// 启用：stack.WithDebug(":6060")
// 模块注册：debug.Register[*Player]("player", repo, cmdBus)
//
// @title           game-stack Debug API
// @description     开发调试接口：运行时查询聚合数据、执行命令、直接修改内存字段。
// @description     默认监听 127.0.0.1:6060，仅开发环境启用。
// @version         1.0.0
// @host            127.0.0.1:6060
package debug
