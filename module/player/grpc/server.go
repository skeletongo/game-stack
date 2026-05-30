// Package grpc 提供 player 模块的 gRPC 服务端和客户端适配。
//
// 服务端：实现 proto PlayerServer 接口，委托给 Repository。
// 客户端：封装 NewClient / NewClientForPlayer，供其他节点调用。
package grpc

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	grpcgo "google.golang.org/grpc"

	"github.com/skeletongo/game-stack/module/player/domain"
)

// server 实现 proto 生成的 PlayerServer 接口。
type server struct {
	UnimplementedPlayerServer
	repo domain.PlayerRepository
}

// GetPlayer 实现 PlayerServer.GetPlayer。
func (s *server) GetPlayer(ctx context.Context, req *GetPlayerReq) (*GetPlayerResp, error) {
	p, err := s.repo.Load(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	return &GetPlayerResp{
		Uid:   p.ID(),
		Name:  p.Nickname().String(),
		Level: p.Level().Int32(),
		Exp:   p.Exp().Int64(),
		Gold:  p.Gold().Int32(),
	}, nil
}

// Register 注册 gRPC 服务到 due 框架。在模块 Init 中调用。
func Register(proxy *node.Proxy, repo domain.PlayerRepository) {
	proxy.AddServiceProvider("player", &Player_ServiceDesc, &server{repo: repo})
}

// NewClient 创建 gRPC 客户端（服务发现，随机负载均衡）。
func NewClient(proxy *node.Proxy) (PlayerClient, error) {
	conn, err := proxy.NewMeshClient("discovery://player")
	if err != nil {
		return nil, err
	}
	cc, ok := conn.Client().(grpcgo.ClientConnInterface)
	if !ok {
		return nil, err
	}
	return NewPlayerClient(cc), nil
}

// NewClientForPlayer 创建指向玩家所在节点的 gRPC 客户端。
func NewClientForPlayer(proxy *node.Proxy, uid int64) (PlayerClient, error) {
	nid, err := proxy.LocateNode(context.Background(), uid, proxy.GetName())
	if err != nil {
		return nil, err
	}
	conn, err := proxy.NewMeshClient("direct://" + nid)
	if err != nil {
		return nil, err
	}
	cc, ok := conn.Client().(grpcgo.ClientConnInterface)
	if !ok {
		return nil, err
	}
	return NewPlayerClient(cc), nil
}
