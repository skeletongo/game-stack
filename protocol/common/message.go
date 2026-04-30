// Package common 定义游戏框架通用数据类型。
package common

import "time"

// Vector3 三维向量。
type Vector3 struct {
	X float32 `json:"x" msgpack:"x"`
	Y float32 `json:"y" msgpack:"y"`
	Z float32 `json:"z" msgpack:"z"`
}

// Position 世界位置（坐标+朝向+场景ID）。
type Position struct {
	Coordinates *Vector3 `json:"coordinates" msgpack:"coordinates"`
	Rotation    float32  `json:"rotation" msgpack:"rotation"`
	SceneID     string   `json:"sceneId" msgpack:"sceneId"`
}

// Currency 通用货币类型。
type Currency struct {
	Gold    int32 `json:"gold" msgpack:"gold"`
	Diamond int32 `json:"diamond" msgpack:"diamond"`
	Stamina int32 `json:"stamina" msgpack:"stamina"`
}

// TimeRange 时间区间。
type TimeRange struct {
	StartAt time.Time `json:"startAt" msgpack:"startAt"`
	EndAt   time.Time `json:"endAt" msgpack:"endAt"`
}

// Pagination 分页参数。
type Pagination struct {
	Page     int32 `json:"page" msgpack:"page"`
	PageSize int32 `json:"pageSize" msgpack:"pageSize"`
	Total    int32 `json:"total" msgpack:"total"`
}
