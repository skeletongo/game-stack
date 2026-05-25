package player

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"google.golang.org/grpc"
)

// 本文件封装 proto 自动生成的 gRPC 代码与 due 框架的对接。
// player.pb.go 和 player_grpc.pb.go 由 protoc 自动生成，
// 使用根目录 gen_proto.sh 一次生成所有模块的 proto 代码。

// server 实现 proto 生成的 PlayerServer 接口，委托给本地 Service。
type server struct {
	UnimplementedPlayerServer
	svc Service
}

// GetPlayer 实现 proto 生成的 PlayerServer.GetPlayer。
func (s *server) GetPlayer(ctx context.Context, req *GetPlayerReq) (*GetPlayerResp, error) {
	p, err := s.svc.GetPlayer(req.Uid)
	if err != nil {
		return nil, err
	}
	return &GetPlayerResp{
		Uid:   p.ID,
		Name:  p.Nickname,
		Level: p.Level,
		Exp:   p.Exp,
		Gold:  p.Gold,
	}, nil
}

// RegisterGRPC 注册 gRPC 服务到 due 框架。在模块 Init 中调用。
func RegisterGRPC(proxy *node.Proxy, svc Service) {
	proxy.AddServiceProvider("player", &Player_ServiceDesc, &server{svc: svc})
}

// NewClient 创建 gRPC 客户端（服务发现，随机负载均衡）。
// 适用于无状态服务或全局查询。
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
// 先通过 LocateNode 定位 uid 的绑定节点，再直连。
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
