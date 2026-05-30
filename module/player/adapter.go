package player

import (
	"context"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/player/application"
	"github.com/skeletongo/game-stack/module/player/domain"
)

// ---- 框架适配器 ----
//
// 适配器是模块与框架之间的胶水层：
//   - cleanableAdapter → 适配 Repository 到框架的生命周期接口（CleanableService）
//   - svcAdapter       → 适配 Repository + CommandBus 到框架的服务注册接口（跨模块调用）

// cleanableAdapter 适配 PlayerRepository → stack.CleanableService。
type cleanableAdapter struct {
	repo domain.PlayerRepository
}

// CleanPlayerData 清理玩家内存数据（断线 Grace Period 到期时调用）。
func (a *cleanableAdapter) CleanPlayerData(uid int64) error {
	return a.repo.Delete(context.Background(), uid)
}

// svcAdapter 是 player 模块对外的服务适配器。
// 其他模块通过 stack.GetService("player") 获取，类型断言为 *svcAdapter。
//
// 提供的能力：
//   - GetPlayer(id) → 查询玩家信息
//   - AddExp / AddGold / DeductGold / AddDiamond / DeductDiamond → 修改玩家资源
type svcAdapter struct {
	repo   domain.PlayerRepository
	cmdBus *ddd.CommandBus
}

// GetPlayer 查询玩家（直接读仓储，不走 Actor）。
func (s *svcAdapter) GetPlayer(id int64) (*domain.Player, error) {
	return s.repo.Load(context.Background(), id)
}

// AddExp 增加经验值（通过命令总线，需在 Actor 内调用）。
func (s *svcAdapter) AddExp(id int64, exp int64) error {
	return s.cmdBus.Dispatch(context.Background(), application.AddExpCmd{PlayerID: id, Amount: exp})
}

// AddGold 增加金币。
func (s *svcAdapter) AddGold(id int64, gold int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.AddGoldCmd{PlayerID: id, Amount: gold})
}

// DeductGold 扣除金币。
func (s *svcAdapter) DeductGold(id int64, gold int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.DeductGoldCmd{PlayerID: id, Amount: gold})
}

// AddDiamond 增加钻石。
func (s *svcAdapter) AddDiamond(id int64, diamond int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.AddDiamondCmd{PlayerID: id, Amount: diamond})
}

// DeductDiamond 扣除钻石。
func (s *svcAdapter) DeductDiamond(id int64, diamond int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.DeductDiamondCmd{PlayerID: id, Amount: diamond})
}
