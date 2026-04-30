package match

import (
	"fmt"
	"time"
)

// Service 匹配模块对外服务接口。
type Service interface {
	// GetQueueSize 获取指定模式的匹配队列大小。
	GetQueueSize(matchType, mode string) int32
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) GetQueueSize(matchType, mode string) int32 {
	// 简化实现
	return 0
}

// winMatch 简单匹配算法：找到 rank 在范围内的第一个对手。
func winMatch(seeker *MatchInfo, candidates []*MatchInfo) *MatchInfo {
	for _, c := range candidates {
		if c.UID == seeker.UID {
			continue
		}
		if c.MatchType == seeker.MatchType && c.Mode == seeker.Mode {
			if c.Rank >= seeker.MinRank && c.Rank <= seeker.MaxRank &&
				seeker.Rank >= c.MinRank && seeker.Rank <= c.MaxRank {
				return c
			}
		}
	}
	return nil
}

func generateMatchID() string {
	return fmt.Sprintf("match_%d", time.Now().UnixNano())
}
