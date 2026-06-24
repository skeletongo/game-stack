package server

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"

	"github.com/skeletongo/game-stack/module/player/internal/domain"
	pb "github.com/skeletongo/game-stack/module/player/rpc/grpc"
)

// Register 注册 player RPC 服务提供者。
func Register(name string, proxy *node.Proxy, repo domain.PlayerRepository) {
	proxy.AddServiceProvider(name, &pb.Player_ServiceDesc, &server{repo: repo})
}

// server 实现 proto 生成的 PlayerServer 接口。
type server struct {
	pb.UnimplementedPlayerServer
	repo domain.PlayerRepository // 玩家仓储
}

// GetPlayer 查询玩家信息。
func (s *server) GetPlayer(ctx context.Context, req *pb.GetPlayerReq) (*pb.GetPlayerResp, error) {
	p, err := s.repo.Load(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	return &pb.GetPlayerResp{
		Uid:       p.ID(),
		Name:      p.Nickname().String(),
		Level:     p.Level().Int32(),
		Exp:       p.Exp().Int64(),
		Gold:      p.Gold().Int32(),
		Diamond:   p.Diamond().Int32(),
		Avatar:    p.Avatar().String(),
		CreatedAt: p.CreatedAt(),
	}, nil
}
