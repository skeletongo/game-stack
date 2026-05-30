package main

import (
	"fmt"

	"github.com/dobyte/due/network/ws/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/client"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/proto/auth"
	"github.com/skeletongo/game-stack/stack"
)

func main() {
	container := due.NewContainer()

	c := client.NewClient(
		client.WithClient(ws.NewClient()),
	)

	// 保存凭据，注册成功后自动登录
	var creds = struct {
		username string
		password string
	}{}

	// 监听组件启动
	c.Proxy().AddHookListener(cluster.Start, func(proxy *client.Proxy) {
		// 拨号连接到 gate（地址从 etc/etc.yaml 的 network.ws.client.url 读取）
		conn, err := c.Proxy().Dial()
		if err != nil {
			log.Fatalf("dial failed: %v", err)
		}
		log.Infof("dialing gate... conn_id=%d", conn.ID())
	})

	// 监听注册响应 → 自动登录
	c.Proxy().AddRouteHandler(stack.RouteAuthRegister, func(ctx *client.Context) {
		resp := &auth.RegisterResponse{}
		if err := ctx.Parse(resp); err != nil {
			log.Errorf("parse register resp failed: %v", err)
			return
		}
		log.Infof("<<< [register] ok: player_id=%d token=%s", resp.PlayerId, resp.Token)

		// 注册成功 → 登录
		log.Infof(">>> [login] username=%s", creds.username)
		loginReq := &auth.LoginRequest{
			Username: creds.username,
			Password: creds.password,
		}
		if err := ctx.Conn().Push(&cluster.Message{
			Route: stack.RouteAuthLogin,
			Data:  loginReq,
		}); err != nil {
			log.Fatalf("login push failed: %v", err)
		}
	})

	// 监听登录响应
	c.Proxy().AddRouteHandler(stack.RouteAuthLogin, func(ctx *client.Context) {
		resp := &auth.LoginResponse{}
		if err := ctx.Parse(resp); err != nil {
			log.Errorf("parse login resp failed: %v", err)
			return
		}
		log.Infof("<<< [login] ok: player_id=%d token=%s is_new=%v",
			resp.PlayerId, resp.Token, resp.IsNewPlayer)
		log.Infof("=== test complete ===")
	})

	// 连接成功后注册
	c.Proxy().AddEventListener(cluster.Connect, func(conn *client.Conn) {
		log.Infof("=== connected to gate ===")

		creds.username = fmt.Sprintf("test_%d", conn.ID())
		creds.password = "123456"

		log.Infof(">>> [register] username=%s", creds.username)
		regReq := &auth.RegisterRequest{
			Username: creds.username,
			Password: creds.password,
			Nickname: "测试玩家",
		}
		if err := conn.Push(&cluster.Message{
			Route: stack.RouteAuthRegister,
			Data:  regReq,
		}); err != nil {
			log.Fatalf("register push failed: %v", err)
		}
	})

	container.Add(c)
	container.Serve()
}
