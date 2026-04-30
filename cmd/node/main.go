package main

import (
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/etcd/v2"
	"github.com/dobyte/due/transport/grpc/v2"

	"github.com/skeletongo/game-stack/module/activity"
	"github.com/skeletongo/game-stack/module/auth"
	"github.com/skeletongo/game-stack/module/chat"
	"github.com/skeletongo/game-stack/module/combat"
	"github.com/skeletongo/game-stack/module/guild"
	"github.com/skeletongo/game-stack/module/inventory"
	"github.com/skeletongo/game-stack/module/leaderboard"
	"github.com/skeletongo/game-stack/module/mail"
	"github.com/skeletongo/game-stack/module/match"
	"github.com/skeletongo/game-stack/module/player"
	"github.com/skeletongo/game-stack/module/quest"
	"github.com/skeletongo/game-stack/module/room"
	"github.com/skeletongo/game-stack/module/shop"
	"github.com/skeletongo/game-stack/module/social"
	"github.com/skeletongo/game-stack/stack"
)

func main() {
	app := stack.NewApplication(
		stack.WithName("game-node"),
		stack.WithLocator(redis.NewLocator()),
		stack.WithRegistry(etcd.NewRegistry()),
		stack.WithTransporter(grpc.NewTransporter()),
		stack.WithModules(
			auth.Module(),
			player.Module(),
			mail.Module(),
			chat.Module(),
			match.Module(),
			room.Module(),
			inventory.Module(),
			quest.Module(),
			combat.Module(),
			guild.Module(),
			shop.Module(),
			leaderboard.Module(),
			activity.Module(),
			social.Module(),
		),
	)

	app.Run()
}
