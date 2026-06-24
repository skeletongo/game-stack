package context

import (
	"fmt"
	"sync"

	"github.com/dobyte/due/component/http/v2"
	"github.com/dobyte/due/v2/cluster/node"

	"github.com/skeletongo/game-stack/cmd/frontend/app/middleware"
	"github.com/skeletongo/game-stack/internal/component/jwt"
	authrpcclient "github.com/skeletongo/game-stack/module/auth/rpc/client"
	authrpc "github.com/skeletongo/game-stack/module/auth/rpc/grpc"
)

var once = new(sync.Once)
var svcCtx *ServiceContext

// ServiceContext 保存 frontend 运行期依赖。
type ServiceContext struct {
	ProxyHttp *http.Proxy      // HTTP 代理
	ProxyNode *node.Proxy      // 节点代理
	Auth      *middleware.Auth // 鉴权中间件

	mu      sync.Mutex         // RPC 客户端缓存锁
	AuthRPC authrpc.AuthClient // auth RPC 客户端
}

// NewServiceContext 初始化全局服务上下文。
func NewServiceContext(proxyHttp *http.Proxy, proxyNode *node.Proxy) *ServiceContext {
	once.Do(func() {
		c := &ServiceContext{
			ProxyHttp: proxyHttp,
			ProxyNode: proxyNode,
			Auth:      middleware.NewAuth(proxyHttp, jwt.Instance()),
		}
		svcCtx = c
		// 统一加日志
		svcCtx.ProxyHttp.Router().Group("/", middleware.Log)
	})

	return svcCtx
}

// AuthClient 获取 auth RPC 客户端。
func (c *ServiceContext) AuthClient() (authrpc.AuthClient, error) {
	if c == nil || c.ProxyNode == nil {
		return nil, fmt.Errorf("frontend service context is not initialized")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.AuthRPC != nil {
		return c.AuthRPC, nil
	}
	cli, err := authrpcclient.New(c.ProxyNode)
	if err != nil {
		return nil, err
	}
	c.AuthRPC = cli
	return cli, nil
}

// GetSvc 返回全局服务上下文。
func GetSvc() *ServiceContext {
	return svcCtx
}
