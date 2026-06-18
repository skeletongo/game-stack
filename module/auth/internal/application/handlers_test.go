package application

import (
	"context"
	"testing"
	"time"

	"github.com/dobyte/due/v2/errors"
	"github.com/skeletongo/game-stack/component/jwt"
	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	"github.com/skeletongo/game-stack/module/auth/internal/infrastructure"
	"github.com/skeletongo/game-stack/stack"
)

func TestRegisterHandlerCreatesAccountAndPlayer(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewMemoryRepo()
	events := ddd.NewEventBus()
	players := newFakePlayerService()
	var created []ddd.DomainEvent
	events.Subscribe(domain.EventAccountCreated, func(event ddd.DomainEvent) {
		created = append(created, event)
	})

	handler := &RegisterHandler{Repo: repo, EventBus: events, Players: players}
	result, err := handler.Handle(ctx, RegisterCmd{
		Username: "alice",
		Password: "secret",
		Nickname: "Alice",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if result.UserID == 0 || result.PlayerID != result.UserID {
		t.Fatalf("unexpected register result: %+v", result)
	}
	if _, err := repo.FindByUsername(ctx, "alice"); err != nil {
		t.Fatalf("account was not saved: %v", err)
	}
	if got := players.created[result.PlayerID]; got != "Alice" {
		t.Fatalf("player was not created with nickname: got %q", got)
	}
	if len(created) != 1 || created[0].AggregateID() != result.UserID {
		t.Fatalf("account created event mismatch: %+v", created)
	}

	_, err = handler.Handle(ctx, RegisterCmd{
		Username: "alice",
		Password: "secret",
		Nickname: "Alice2",
	})
	if !errors.Is(err, stack.ErrAccountExists) {
		t.Fatalf("duplicate username error = %v, want %v", err, stack.ErrAccountExists)
	}
}

func TestLoginHandlerRejectsWrongPassword(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewMemoryRepo()
	account := mustNewAccount(t, 1001, "alice", "secret", "Alice")
	if err := repo.Save(ctx, account); err != nil {
		t.Fatalf("save account: %v", err)
	}

	handler := &LoginHandler{Repo: repo, EventBus: ddd.NewEventBus(), Jwt: newTestJWT()}
	_, err := handler.Handle(ctx, LoginCmd{Username: "alice", Password: "bad"})
	if !errors.Is(err, stack.ErrWrongPassword) {
		t.Fatalf("login error = %v, want %v", err, stack.ErrWrongPassword)
	}
}

func TestTokenLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewMemoryRepo()
	events := ddd.NewEventBus()
	jt := newTestJWT()
	account := mustNewAccount(t, 1001, "alice", "secret", "Alice")
	if err := repo.Save(ctx, account); err != nil {
		t.Fatalf("save account: %v", err)
	}

	login := &LoginHandler{Repo: repo, EventBus: events, Jwt: jt}
	loginResult, err := login.Handle(ctx, LoginCmd{Username: "alice", Password: "secret"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResult.Token == "" {
		t.Fatal("login returned empty token")
	}
	if loginResult.ExpiresAt <= time.Now().Unix() {
		t.Fatalf("expires_at = %d, want future timestamp", loginResult.ExpiresAt)
	}
	payload, err := jt.ParseToken(loginResult.Token)
	if err != nil {
		t.Fatalf("parse login token: %v", err)
	}
	if payload.Subject() != "1001" {
		t.Fatalf("token subject = %q, want 1001", payload.Subject())
	}

	markOnline := &MarkOnlineHandler{Repo: repo, EventBus: events}
	if _, err := markOnline.Handle(ctx, MarkOnlineCmd{UserID: loginResult.UserID, Token: loginResult.Token, GID: "gate-1"}); err != nil {
		t.Fatalf("mark online failed: %v", err)
	}
	onlineAccount, err := repo.FindByToken(ctx, loginResult.Token)
	if err != nil {
		t.Fatalf("token index was not saved: %v", err)
	}
	if !onlineAccount.IsOnline() || onlineAccount.OnlineGID() != "gate-1" {
		t.Fatalf("account online state mismatch: online=%v gid=%q", onlineAccount.IsOnline(), onlineAccount.OnlineGID())
	}

	time.Sleep(time.Millisecond)

	refresh := &RefreshTokenHandler{Repo: repo, Jwt: jt}
	refreshResult, err := refresh.Handle(ctx, RefreshTokenCmd{UserID: loginResult.UserID, Token: loginResult.Token})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshResult.Token == "" || refreshResult.Token == loginResult.Token {
		t.Fatalf("refresh token mismatch: old=%q new=%q", loginResult.Token, refreshResult.Token)
	}
	refreshedAccount, err := repo.Load(ctx, loginResult.UserID)
	if err != nil {
		t.Fatalf("load refreshed account: %v", err)
	}
	if refreshedAccount.Token().String() != refreshResult.Token {
		t.Fatalf("account token = %q, want refreshed token", refreshedAccount.Token().String())
	}
	_, err = refresh.Handle(ctx, RefreshTokenCmd{UserID: loginResult.UserID, Token: loginResult.Token})
	if !errors.Is(err, stack.ErrInvalidToken) {
		t.Fatalf("refresh old token error = %v, want %v", err, stack.ErrInvalidToken)
	}

	logout := &LogoutHandler{Repo: repo, EventBus: events, Jwt: jt}
	if _, err := logout.Handle(ctx, LogoutCmd{UserID: loginResult.UserID}); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	loggedOutAccount, err := repo.Load(ctx, loginResult.UserID)
	if err != nil {
		t.Fatalf("load logged out account: %v", err)
	}
	if !loggedOutAccount.Token().IsEmpty() || loggedOutAccount.IsOnline() {
		t.Fatalf("account was not logged out: token=%q online=%v", loggedOutAccount.Token().String(), loggedOutAccount.IsOnline())
	}
	if _, err := repo.FindByToken(ctx, refreshResult.Token); err == nil {
		t.Fatal("refreshed token should be removed from repository index after logout")
	}
}

func newTestJWT() *jwt.JWT {
	return jwt.NewInstance(jwt.Config{
		Issuer:          "auth-test",
		AudienceKey:     "game-stack-test",
		SecretKey:       "test-secret",
		ValidDuration:   time.Hour,
		RefreshDuration: 2 * time.Hour,
	})
}

func mustNewAccount(t *testing.T, id int64, username, password, nickname string) *domain.Account {
	t.Helper()
	account, err := domain.NewAccount(id, id, username, password, nickname)
	if err != nil {
		t.Fatalf("new account: %v", err)
	}
	return account
}

type fakePlayerService struct {
	created map[int64]string
	deleted map[int64]bool
}

func newFakePlayerService() *fakePlayerService {
	return &fakePlayerService{
		created: make(map[int64]string),
		deleted: make(map[int64]bool),
	}
}

func (f *fakePlayerService) CreatePlayer(_ context.Context, id int64, nickname string) error {
	f.created[id] = nickname
	return nil
}

func (f *fakePlayerService) DeletePlayer(_ context.Context, id int64) error {
	delete(f.created, id)
	f.deleted[id] = true
	return nil
}
