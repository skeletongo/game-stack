package logfmt

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const redacted = "***"

// JSON returns compact JSON with sensitive fields redacted.
func JSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"type":%q,"error":%q}`, fmt.Sprintf("%T", v), "marshal failed")
	}
	return redactJSON(b)
}

// ProtoJSON returns compact proto JSON with sensitive fields redacted.
func ProtoJSON(m proto.Message) string {
	if m == nil {
		return "null"
	}
	b, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}.Marshal(m)
	if err != nil {
		return fmt.Sprintf(`{"type":%q,"error":%q}`, fmt.Sprintf("%T", m), "marshal failed")
	}
	return redactJSON(b)
}

func redactJSON(b []byte) string {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	redact(v, "")
	out, err := json.Marshal(v)
	if err != nil {
		return string(b)
	}
	return string(out)
}

func redact(v any, field string) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if isSensitive(k) {
				x[k] = redacted
				continue
			}
			redact(child, k)
		}
	case []any:
		for _, child := range x {
			redact(child, field)
		}
	}
}

func isSensitive(field string) bool {
	name := strings.ToLower(field)
	return strings.Contains(name, "password") ||
		strings.Contains(name, "token") ||
		strings.Contains(name, "secret") ||
		strings.Contains(name, "key")
}
