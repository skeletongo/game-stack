// Package activity 定义活动系统相关消息类型。
package activity

// ActivityInfo 活动信息。
type ActivityInfo struct {
	ID          int32   `json:"id" msgpack:"id"`
	Name        string  `json:"name" msgpack:"name"`
	Description string  `json:"description" msgpack:"description"`
	Type        int32   `json:"type" msgpack:"type"`     // 1:签到 2:累计 3:限时 4:BOSS
	Status      int32   `json:"status" msgpack:"status"` // 0:未开启 1:进行中 2:已结束
	StartAt     int64   `json:"startAt" msgpack:"startAt"`
	EndAt       int64   `json:"endAt" msgpack:"endAt"`
	Progress    int32   `json:"progress" msgpack:"progress"`
	Target      int32   `json:"target" msgpack:"target"`
	Claimed     bool    `json:"claimed" msgpack:"claimed"`
	Rewards     []int32 `json:"rewards" msgpack:"rewards"`
}

// ListRequest 获取活动列表请求。
type ListRequest struct {
	Type int32 `json:"type" msgpack:"type"` // -1:全部
}

// ListResponse 获取活动列表响应。
type ListResponse struct {
	Activities []*ActivityInfo `json:"activities" msgpack:"activities"`
}

// ClaimRequest 领取活动奖励请求。
type ClaimRequest struct {
	ActivityID int32 `json:"activityId" msgpack:"activityId"`
}

// ClaimResponse 领取活动奖励响应。
type ClaimResponse struct {
	Rewards []int32 `json:"rewards" msgpack:"rewards"`
}

// ProgressEvent 活动进度事件（服务器推送）。
type ProgressEvent struct {
	ActivityID int32 `json:"activityId" msgpack:"activityId"`
	Progress   int32 `json:"progress" msgpack:"progress"`
	Target     int32 `json:"target" msgpack:"target"`
}

// StartEvent 活动开始事件（服务器推送）。
type StartEvent struct {
	ActivityID int32 `json:"activityId" msgpack:"activityId"`
	EndAt      int64 `json:"endAt" msgpack:"endAt"`
}
