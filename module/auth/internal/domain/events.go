package domain

import "time"

// 领域事件名称常量。
const (
	EventAccountCreated   = "auth.account_created"
	EventAccountLoggedIn  = "auth.account_logged_in"
	EventAccountLoggedOut = "auth.account_logged_out"
	EventAccountBanned    = "auth.account_banned"
)

// AccountCreated 账户创建事件。
type AccountCreated struct {
	userID     int64
	username   string
	occurredAt time.Time
}

func NewAccountCreated(userID int64, username string) AccountCreated {
	return AccountCreated{userID: userID, username: username, occurredAt: time.Now()}
}
func (e AccountCreated) AggregateID() int64    { return e.userID }
func (e AccountCreated) EventName() string     { return EventAccountCreated }
func (e AccountCreated) OccurredAt() time.Time { return e.occurredAt }

// AccountLoggedIn 账户登录事件。
type AccountLoggedIn struct {
	userID     int64
	occurredAt time.Time
}

func NewAccountLoggedIn(userID int64) AccountLoggedIn {
	return AccountLoggedIn{userID: userID, occurredAt: time.Now()}
}
func (e AccountLoggedIn) AggregateID() int64    { return e.userID }
func (e AccountLoggedIn) EventName() string     { return EventAccountLoggedIn }
func (e AccountLoggedIn) OccurredAt() time.Time { return e.occurredAt }

// AccountLoggedOut 账户登出事件。
type AccountLoggedOut struct {
	userID     int64
	occurredAt time.Time
}

func NewAccountLoggedOut(userID int64) AccountLoggedOut {
	return AccountLoggedOut{userID: userID, occurredAt: time.Now()}
}
func (e AccountLoggedOut) AggregateID() int64    { return e.userID }
func (e AccountLoggedOut) EventName() string     { return EventAccountLoggedOut }
func (e AccountLoggedOut) OccurredAt() time.Time { return e.occurredAt }

// AccountBanned 账户封禁事件。
type AccountBanned struct {
	userID     int64
	occurredAt time.Time
}

func NewAccountBanned(userID int64) AccountBanned {
	return AccountBanned{userID: userID, occurredAt: time.Now()}
}
func (e AccountBanned) AggregateID() int64    { return e.userID }
func (e AccountBanned) EventName() string     { return EventAccountBanned }
func (e AccountBanned) OccurredAt() time.Time { return e.occurredAt }
