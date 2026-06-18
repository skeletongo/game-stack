package logfmt

import (
	"strings"
	"testing"

	"github.com/skeletongo/game-stack/proto/auth"
)

func TestJSONRedactsSensitiveFields(t *testing.T) {
	type nested struct {
		SecretKey string `json:"secret_key"`
		Visible   string `json:"visible"`
	}
	payload := struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		AuthToken string `json:"auth_token"`
		Nested    nested `json:"nested"`
	}{
		Username:  "alice",
		Password:  "secret",
		AuthToken: "token-value",
		Nested:    nested{SecretKey: "key-value", Visible: "ok"},
	}

	got := JSON(payload)
	for _, leaked := range []string{"\"secret\"", "token-value", "key-value"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("payload leaked %q: %s", leaked, got)
		}
	}
	for _, want := range []string{`"username":"alice"`, `"password":"***"`, `"auth_token":"***"`, `"secret_key":"***"`, `"visible":"ok"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("payload missing %s: %s", want, got)
		}
	}
}

func TestProtoJSONRedactsSensitiveFields(t *testing.T) {
	got := ProtoJSON(&auth.LoginResponse{
		Code:      0,
		Token:     "token-value",
		PlayerId:  1001,
		ExpiresAt: 2002,
	})
	if strings.Contains(got, "token-value") {
		t.Fatalf("proto payload leaked token: %s", got)
	}
	for _, want := range []string{`"token":"***"`, `"player_id":"1001"`, `"expires_at":"2002"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("proto payload missing %s: %s", want, got)
		}
	}
}

func TestProtoJSONNil(t *testing.T) {
	if got := ProtoJSON(nil); got != "null" {
		t.Fatalf("ProtoJSON(nil) = %s, want null", got)
	}
}
