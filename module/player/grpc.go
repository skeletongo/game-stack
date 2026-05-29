package player

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"google.golang.org/grpc"

	"github.com/skeletongo/game-stack/module/player/domain"
)

// server 实现 proto 生成的 PlayerServer 接口，委托给 Repository。
type server struct {
	UnimplementedPlayerServer
	repo domain.PlayerRepository
}

// GetPlayer 实现 proto PlayerServer.GetPlayer。
// 直接通过仓储查询（gRPC 调用来自其他节点，不走 Actor）。
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

// RegisterGRPC 注册 gRPC 服务到 due 框架。在模块 Init 中调用。
func RegisterGRPC(proxy *node.Proxy, repo domain.PlayerRepository) {
	proxy.AddServiceProvider("player", &Player_ServiceDesc, &server{repo: repo})
}

// NewClient 创建 gRPC 客户端（服务发现，随机负载均衡）。
// 适用于无状态查询。
func NewClient(proxy *node.Proxy) (PlayerClient, error) {
	conn, err := proxy.NewMeshClient("discovery://player")
	if err != nil {
		return nil, err
	}
	cc, ok := conn.Client().(grpc.ClientConnInterface)
	if !ok {
		return nil, err
	}
	return NewPlayerClient(cc), nil
}

// NewClientForPlayer 创建指向玩家所在节点的 gRPC 客户端。
// 先通过 LocateNode 定位玩家的绑定节点，再直连。
// 适用于需要访问玩家内存数据的操作。
func NewClientForPlayer(proxy *node.Proxy, uid int64) (PlayerClient, error) {
	nid, err := proxy.LocateNode(context.Background(), uid, proxy.GetName())
	if err != nil {
		return nil, err
	}
	conn, err := proxy.NewMeshClient("direct://" + nid)
	if err != nil {
		return nil, err
	}
	cc, ok := conn.Client().(grpc.ClientConnInterface)
	if !ok {
		return nil, err
	}
	return NewPlayerClient(cc), nil
}
