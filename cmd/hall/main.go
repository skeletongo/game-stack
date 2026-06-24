package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/etcd/v2"
	"github.com/dobyte/due/transport/grpc/v2"

	"github.com/skeletongo/game-stack/module/auth"
	"github.com/skeletongo/game-stack/module/player"
	"github.com/skeletongo/game-stack/stack"
)

func main() {
	app := stack.NewApplication(
		stack.WithName("game-hall"),
		stack.WithLocator(redis.NewLocator()),
		stack.WithRegistry(etcd.NewRegistry()),
		stack.WithTransporter(grpc.NewTransporter()),
		stack.WithDebug("127.0.0.1:6060"),
		stack.WithModules(
			auth.Module(),
			player.Module(),
		),
	)

	app.Run()
}
