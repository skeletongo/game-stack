package application

import (
	"github.com/skeletongo/game-stack/module/player/internal/domain"
	"github.com/skeletongo/game-stack/proto/player"
)

// PlayerToProto 将领域 Player 聚合转换为 proto PlayerInfo 消息。
func PlayerToProto(p *domain.Player) *player.PlayerInfo {
	return &player.PlayerInfo{
		Id:        p.ID(),
		Nickname:  p.Nickname().String(),
		Level:     p.Level().Int32(),
		Exp:       p.Exp().Int64(),
		Avatar:    p.Avatar().String(),
		Gold:      p.Gold().Int32(),
		Diamond:   p.Diamond().Int32(),
		CreatedAt: p.CreatedAt(),
	}
}
