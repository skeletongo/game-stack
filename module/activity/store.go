package activity

import "context"

type ActivityInfo struct {
	ID          int32
	Name        string
	Description string
	Type        int32
	Status      int32
	StartAt     int64
	EndAt       int64
	Progress    int32
	Target      int32
	Claimed     bool
	Rewards     []int32
}

type Store interface {
	GetActivity(ctx context.Context, activityID int32) (*ActivityInfo, error)
	ListActivities(ctx context.Context, aType int32) ([]*ActivityInfo, error)
	ClaimReward(ctx context.Context, uid int64, activityID int32) error
	UpdateProgress(ctx context.Context, uid int64, activityID int32, progress int32) error
}
