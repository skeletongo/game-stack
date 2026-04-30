package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/network/ws/v2"
	"github.com/dobyte/due/registry/etcd/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/gate"
)

func main() {
	container := due.NewContainer()

	g := gate.NewGate(
		gate.WithName("gate"),
		gate.WithServer(ws.NewServer()),
		gate.WithLocator(redis.NewLocator()),
		gate.WithRegistry(etcd.NewRegistry()),
	)

	container.Add(g)
	container.Serve()
}
