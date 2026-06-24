package client

import (
	"context"
	"fmt"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/transport"
	"google.golang.org/grpc"

	pb "github.com/skeletongo/game-stack/module/player/rpc/grpc"
)

// newClientFromConn 从 MeshClient 连接创建 PlayerClient。
func newClientFromConn(conn transport.Client) (pb.PlayerClient, error) {
	cc, ok := conn.Client().(grpc.ClientConnInterface)
	if !ok {
		return nil, fmt.Errorf("mesh client does not implement grpc.ClientConnInterface")
	}
	return pb.NewPlayerClient(cc), nil
}

// New 创建按服务发现寻址的 player RPC 客户端。
func New(proxy *node.Proxy) (pb.PlayerClient, error) {
	conn, err := proxy.NewMeshClient("discovery://player")
	if err != nil {
		return nil, err
	}
	return newClientFromConn(conn)
}

// NewWithUID 创建按玩家节点归属直连的 player RPC 客户端。
func NewWithUID(ctx context.Context, proxy *node.Proxy, nodeName string, uid int64) (pb.PlayerClient, error) {
	nid, err := proxy.LocateNode(ctx, uid, nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := proxy.NewMeshClient("direct://" + nid)
	if err != nil {
		return nil, err
	}
	return newClientFromConn(conn)
}
