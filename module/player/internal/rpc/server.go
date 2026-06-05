// Package rpc 提供 player 模块的 RPC 服务端和客户端适配。
//
// 服务端：实现 proto PlayerServer 接口，委托给 Repository。
// 客户端：封装 NewClient / NewClientForPlayer，供其他节点调用。
package rpc

import (
	"context"
	"fmt"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/transport"
	grpcgo "google.golang.org/grpc"

	pb "github.com/skeletongo/game-stack/module/player/grpc"
	"github.com/skeletongo/game-stack/module/player/internal/domain"
)

// server 实现 proto 生成的 PlayerServer 接口。
type server struct {
	pb.UnimplementedPlayerServer
	repo domain.PlayerRepository
}

// GetPlayer 实现 PlayerServer.GetPlayer。
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

// Register 注册 RPC 服务到 due 框架。在模块 Init 中调用。
func Register(name string, proxy *node.Proxy, repo domain.PlayerRepository) {
	proxy.AddServiceProvider(name, &pb.Player_ServiceDesc, &server{repo: repo})
}

// NewClient 创建 RPC 客户端（服务发现，随机负载均衡）。
func NewClient(proxy *node.Proxy) (pb.PlayerClient, error) {
	conn, err := proxy.NewMeshClient("discovery://player")
	if err != nil {
		return nil, err
	}
	return newClientFromConn(conn)
}

// NewClientForPlayer 创建指向玩家所在节点的 RPC 客户端。
func NewClientForPlayer(proxy *node.Proxy, uid int64) (pb.PlayerClient, error) {
	nid, err := proxy.LocateNode(context.Background(), uid, proxy.GetName())
	if err != nil {
		return nil, err
	}
	conn, err := proxy.NewMeshClient("direct://" + nid)
	if err != nil {
		return nil, err
	}
	return newClientFromConn(conn)
}

// newClientFromConn 从 MeshClient 连接创建 PlayerClient。
func newClientFromConn(conn transport.Client) (pb.PlayerClient, error) {
	cc, ok := conn.Client().(grpcgo.ClientConnInterface)
	if !ok {
		return nil, fmt.Errorf("mesh client does not implement grpc.ClientConnInterface")
	}
	return pb.NewPlayerClient(cc), nil
}
