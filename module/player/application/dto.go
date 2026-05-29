package application

import (
	pb "github.com/skeletongo/game-stack/proto/player"

	"github.com/skeletongo/game-stack/module/player/domain"
)

// PlayerToProto 将领域 Player 聚合转换为 proto PlayerInfo 消息。
func PlayerToProto(p *domain.Player) *pb.PlayerInfo {
	return &pb.PlayerInfo{
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
