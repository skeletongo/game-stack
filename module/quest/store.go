package quest

import "context"

type QuestStatus int32

const (
	QuestLocked  QuestStatus = 0
	QuestOpen    QuestStatus = 1
	QuestDoing   QuestStatus = 2
	QuestDone    QuestStatus = 3
	QuestClaimed QuestStatus = 4
)

type Quest struct {
	ID          int32
	Name        string
	Description string
	Type        int32
	Status      QuestStatus
	Progress    int32
	Target      int32
	RewardGold  int32
	RewardExp   int64
	RewardItems []int32
}

type Store interface {
	GetAllQuests(ctx context.Context) ([]*Quest, error)
	GetPlayerQuest(ctx context.Context, uid int64, questID int32) (*Quest, error)
	ListPlayerQuests(ctx context.Context, uid int64, qType int32) ([]*Quest, error)
	AcceptQuest(ctx context.Context, uid int64, questID int32) error
	SubmitQuest(ctx context.Context, uid int64, questID int32) error
	UpdateProgress(ctx context.Context, uid int64, questID int32, progress int32) error
}
