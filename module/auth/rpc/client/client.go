package client

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/transport"
	"google.golang.org/grpc"

	pb "github.com/skeletongo/game-stack/module/auth/rpc/grpc"
)

// newClientFunc 根据目标地址创建 RPC 客户端。
func newClientFunc[T any](fn transport.NewMeshClient, newFunc func(grpc.ClientConnInterface) T, target string) (T, error) {
	var empty T
	client, err := fn(target)
	if err != nil {
		return empty, err
	}

	return newFunc(client.Client().(grpc.ClientConnInterface)), nil
}

// New 创建按服务发现寻址的 auth RPC 客户端。
func New(proxy *node.Proxy) (pb.AuthClient, error) {
	cli, err := newClientFunc(proxy.NewMeshClient, pb.NewAuthClient, "discovery://auth")
	if err != nil {
		log.Errorf("new auth client failed: %v", err)
		return nil, err
	}
	return cli, nil
}

// NewWithUID 创建按用户节点归属直连的 auth RPC 客户端。
func NewWithUID(ctx context.Context, proxy *node.Proxy, nodeName string, uid int64) (pb.AuthClient, error) {
	nid, err := proxy.LocateNode(ctx, uid, nodeName)
	if err != nil {
		return nil, err
	}
	cli, err := newClientFunc(proxy.NewMeshClient, pb.NewAuthClient, "direct://"+nid)
	if err != nil {
		log.Errorf("new auth client failed: %v", err)
		return nil, err
	}
	return cli, nil
}
