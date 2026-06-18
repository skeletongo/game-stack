package main

import (
	"fmt"
	"sync"
	"time"

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

	type flowState struct {
		mu       sync.Mutex
		username string
		password string
		token    string
		step     string
		done     bool
	}

	var state flowState

	setStep := func(step string) {
		state.mu.Lock()
		state.step = step
		state.mu.Unlock()
	}

	currentStep := func() string {
		state.mu.Lock()
		defer state.mu.Unlock()
		return state.step
	}

	credentials := func() (string, string) {
		state.mu.Lock()
		defer state.mu.Unlock()
		return state.username, state.password
	}

	saveToken := func(token string) {
		state.mu.Lock()
		state.token = token
		state.mu.Unlock()
	}

	currentToken := func() string {
		state.mu.Lock()
		defer state.mu.Unlock()
		return state.token
	}

	failFlow := func(format string, args ...any) {
		log.Errorf(format, args...)
		state.mu.Lock()
		state.done = true
		state.mu.Unlock()
	}

	sendRegister := func(conn *client.Conn, step string) {
		username, password := credentials()
		setStep(step)
		log.Infof(">>> [%s] username=%s", step, username)
		if err := conn.Push(&cluster.Message{
			Route: stack.RouteAuthRegister,
			Data: &auth.RegisterRequest{
				Username: username,
				Password: password,
				Nickname: fmt.Sprintf("test%d", conn.ID()),
			},
		}); err != nil {
			failFlow("%s push failed: %v", step, err)
		}
	}

	sendLogin := func(conn *client.Conn, step string, password string) {
		username, _ := credentials()
		setStep(step)
		log.Infof(">>> [%s] username=%s", step, username)
		if err := conn.Push(&cluster.Message{
			Route: stack.RouteAuthLogin,
			Data: &auth.LoginRequest{
				Username: username,
				Password: password,
			},
		}); err != nil {
			failFlow("%s push failed: %v", step, err)
		}
	}

	sendRefresh := func(conn *client.Conn, step string, token string) {
		setStep(step)
		log.Infof(">>> [%s]", step)
		if err := conn.Push(&cluster.Message{
			Route: stack.RouteAuthTokenRefresh,
			Data:  &auth.TokenRefreshRequest{Token: token},
		}); err != nil {
			failFlow("%s push failed: %v", step, err)
		}
	}

	sendLogout := func(conn *client.Conn) {
		setStep("logout")
		log.Infof(">>> [logout]")
		if err := conn.Push(&cluster.Message{
			Route: stack.RouteAuthLogout,
			Data:  &auth.LogoutRequest{},
		}); err != nil {
			failFlow("logout push failed: %v", err)
		}
	}

	startFlow := func(conn *client.Conn) {
		state.mu.Lock()
		if state.done || state.step != "" {
			state.mu.Unlock()
			return
		}
		state.username = fmt.Sprintf("auth_test_%d_%d", conn.ID(), time.Now().UnixNano())
		state.password = "123456"
		state.mu.Unlock()

		log.Infof("=== auth test start ===")
		sendRegister(conn, "register")
	}

	expectCode := func(step string, got int32, want int32, message string) bool {
		if got != want {
			failFlow("<<< [%s] unexpected code=%d want=%d message=%s", step, got, want, message)
			return false
		}
		if message != "" {
			log.Infof("<<< [%s] code=%d message=%s", step, got, message)
		} else {
			log.Infof("<<< [%s] code=%d", step, got)
		}
		return true
	}

	finishFlow := func() {
		state.mu.Lock()
		state.done = true
		state.step = ""
		state.mu.Unlock()
		log.Infof("=== auth test complete ===")
	}

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
		step := currentStep()
		resp := &auth.RegisterResponse{}
		if err := ctx.Parse(resp); err != nil {
			failFlow("parse register resp failed: %v", err)
			return
		}

		switch step {
		case "register":
			if !expectCode(step, resp.Code, stack.CodeOK, resp.Message) {
				return
			}
			sendRegister(ctx.Conn(), "register_duplicate")
		case "register_duplicate":
			if !expectCode(step, resp.Code, stack.ErrAccountExists.Code, resp.Message) {
				return
			}
			sendLogin(ctx.Conn(), "login_wrong_password", "wrong-password")
		default:
			failFlow("unexpected register response in step=%s code=%d message=%s", step, resp.Code, resp.Message)
		}
	})

	// 监听登录响应
	c.Proxy().AddRouteHandler(stack.RouteAuthLogin, func(ctx *client.Context) {
		step := currentStep()
		resp := &auth.LoginResponse{}
		if err := ctx.Parse(resp); err != nil {
			failFlow("parse login resp failed: %v", err)
			return
		}

		switch step {
		case "login_wrong_password":
			if !expectCode(step, resp.Code, stack.ErrWrongPassword.Code, resp.Message) {
				return
			}
			_, password := credentials()
			sendLogin(ctx.Conn(), "login", password)
		case "login":
			if !expectCode(step, resp.Code, stack.CodeOK, resp.Message) {
				return
			}
			log.Infof("<<< [login] player_id=%d token=%s expires_at=%d is_new=%v",
				resp.PlayerId, resp.Token, resp.ExpiresAt, resp.IsNewPlayer)
			saveToken(resp.Token)
			sendRefresh(ctx.Conn(), "refresh_invalid_token", "bad-token")
		default:
			failFlow("unexpected login response in step=%s code=%d message=%s", step, resp.Code, resp.Message)
		}
	})

	// 监听刷新响应
	c.Proxy().AddRouteHandler(stack.RouteAuthTokenRefresh, func(ctx *client.Context) {
		step := currentStep()
		resp := &auth.TokenRefreshResponse{}
		if err := ctx.Parse(resp); err != nil {
			failFlow("parse token refresh resp failed: %v", err)
			return
		}

		switch step {
		case "refresh_invalid_token":
			if !expectCode(step, resp.Code, stack.ErrInvalidToken.Code, resp.Message) {
				return
			}
			sendRefresh(ctx.Conn(), "refresh", currentToken())
		case "refresh":
			if !expectCode(step, resp.Code, stack.CodeOK, resp.Message) {
				return
			}
			log.Infof("<<< [refresh] token=%s expires_at=%d", resp.Token, resp.ExpiresAt)
			saveToken(resp.Token)
			sendLogout(ctx.Conn())
		default:
			failFlow("unexpected refresh response in step=%s code=%d message=%s", step, resp.Code, resp.Message)
		}
	})

	// 监听登出响应。服务端当前返回空响应体，这里只确认 1003 已收到。
	c.Proxy().AddRouteHandler(stack.RouteAuthLogout, func(ctx *client.Context) {
		log.Infof("<<< [logout] ok: route=%d conn_id=%d", stack.RouteAuthLogout, ctx.Conn().ID())
		finishFlow()
	})

	// 连接成功后执行一次 auth 功能测试
	c.Proxy().AddEventListener(cluster.Connect, func(conn *client.Conn) {
		log.Infof("=== connected to gate ===")
		startFlow(conn)
	})

	container.Add(c)
	container.Serve()
}
