// Package quest 定义任务/成就相关消息类型。
package quest

// Quest 任务信息。
type Quest struct {
	ID          int32   `json:"id" msgpack:"id"`
	Name        string  `json:"name" msgpack:"name"`
	Description string  `json:"description" msgpack:"description"`
	Type        int32   `json:"type" msgpack:"type"`     // 1:主线 2:支线 3:每日 4:成就
	Status      int32   `json:"status" msgpack:"status"` // 0:锁定 1:可接受 2:进行中 3:已完成 4:已领取
	Progress    int32   `json:"progress" msgpack:"progress"`
	Target      int32   `json:"target" msgpack:"target"`
	RewardGold  int32   `json:"rewardGold" msgpack:"rewardGold"`
	RewardExp   int64   `json:"rewardExp" msgpack:"rewardExp"`
	RewardItems []int32 `json:"rewardItems" msgpack:"rewardItems"`
}

// ListRequest 获取任务列表请求。
type ListRequest struct {
	Type int32 `json:"type" msgpack:"type"` // 0:全部
}

// ListResponse 获取任务列表响应。
type ListResponse struct {
	Quests []*Quest `json:"quests" msgpack:"quests"`
}

// AcceptRequest 接受任务请求。
type AcceptRequest struct {
	QuestID int32 `json:"questId" msgpack:"questId"`
}

// SubmitRequest 提交任务请求。
type SubmitRequest struct {
	QuestID int32 `json:"questId" msgpack:"questId"`
}

// SubmitResponse 提交任务响应。
type SubmitResponse struct {
	RewardGold  int32   `json:"rewardGold" msgpack:"rewardGold"`
	RewardExp   int64   `json:"rewardExp" msgpack:"rewardExp"`
	RewardItems []int32 `json:"rewardItems" msgpack:"rewardItems"`
}

// AbandonRequest 放弃任务请求。
type AbandonRequest struct {
	QuestID int32 `json:"questId" msgpack:"questId"`
}

// ProgressEvent 任务进度事件（服务器推送）。
type ProgressEvent struct {
	QuestID  int32 `json:"questId" msgpack:"questId"`
	Progress int32 `json:"progress" msgpack:"progress"`
	Target   int32 `json:"target" msgpack:"target"`
}
