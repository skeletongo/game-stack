package ddd

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

// Snapshot 返回聚合的完整快照。
// 遍历所有非导出字段，遇到 Scalar 类型展开为底层值，time.Time 转为 Unix 秒。
func Snapshot(agg Aggregate) map[string]any {
	v := reflect.ValueOf(agg).Elem()
	t := v.Type()
	m := make(map[string]any, t.NumField())

	for i := range t.NumField() {
		f := t.Field(i)
		if f.IsExported() {
			continue
		}
		fv := v.Field(i)
		m[f.Name] = snapshotValue(fv)
	}
	return m
}

// snapshotValue 递归展开 reflect.Value 为普通 Go 值。
func snapshotValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}

	// Scalar 类型 → 底层值
	if v.CanInterface() {
		if s, ok := v.Interface().(Scalar); ok {
			return s.Scalar()
		}
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return v.Bool()
	case reflect.Struct:
		// time.Time → Unix timestamp
		if v.CanInterface() {
			if t, ok := v.Interface().(time.Time); ok {
				return t.Unix()
			}
		}
		// 其他 struct 递归展开
		sub := make(map[string]any)
		t := v.Type()
		for i := range t.NumField() {
			f := t.Field(i)
			if f.IsExported() {
				continue
			}
			sub[f.Name] = snapshotValue(v.Field(i))
		}
		return sub
	default:
		if v.CanInterface() {
			return v.Interface()
		}
		return fmt.Sprintf("%v", v)
	}
}

// ApplyPatch 直接修改聚合的非导出字段。
// 使用 unsafe 绕过 Go 的导出限制。
// 警告：不经过构造校验，不维护不变量。仅用于调试。
func ApplyPatch(agg Aggregate, fields map[string]any) error {
	v := reflect.ValueOf(agg).Elem()
	t := v.Type()

	for key, val := range fields {
		field := v.FieldByName(key)
		if !field.IsValid() {
			return fmt.Errorf("field %q not found in %s", key, t.Name())
		}
		if err := setField(field, reflect.ValueOf(val)); err != nil {
			return fmt.Errorf("field %q: %w", key, err)
		}
	}
	return nil
}

// setField 用 unsafe 直接写入 reflect.Value。
func setField(field reflect.Value, val reflect.Value) error {
	if field.Kind() == reflect.String && val.Kind() == reflect.String {
		// string 字段：Go 1.20+ 的 reflect 不支持 SetString 给非导出字段。
		// 用 unsafe 绕过。
		ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		ptr.SetString(val.String())
		return nil
	}

	// 数值类型转换
	if val.CanConvert(field.Type()) {
		ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		ptr.Set(val.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot convert %v (%s) to %s", val, val.Type(), field.Type())
}
