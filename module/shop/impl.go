package shop

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	pshop "github.com/skeletongo/game-stack/protocol/shop"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct{ svc *service }

func newImpl(store Store) *impl { return &impl{svc: newService(store)} }

// handleList 获取商城列表。读操作，不需要 Actor。
func (i *impl) handleList(ctx node.Context) {
	req := &pshop.ListRequest{TabIndex: -1}
	_ = ctx.Parse(req)
	items, _ := i.svc.store.ListItems(context.Background(), req.TabIndex)
	resp := &pshop.ListResponse{Items: make([]*pshop.ShopItem, 0, len(items)), TabIndex: req.TabIndex}
	for _, it := range items {
		resp.Items = append(resp.Items, &pshop.ShopItem{
			ID: it.ID, ItemID: it.ItemID, ItemName: it.ItemName,
			Price: it.Price, CurrencyType: it.CurrencyType, Discount: it.Discount,
			LimitCount: it.LimitCount, SoldCount: it.SoldCount, LevelRequire: it.LevelRequire,
		})
	}
	stack.RespondData(ctx, resp)
}

// handleBuyActor 购买商品（模式2：RouteToActor）。
// 此函数在 PlayerActor 的 dispatch goroutine 中串行执行，天然线程安全。
//
// 消息流：
//
//	客户端 → Gate → Node(StatefulRoute) → RouteToActor → Actor.mailbox → 此函数
//
// 关键：ctx.Response() 在此函数中直接调用，是同步的。
// Actor 是单 goroutine 串行处理，所有玩家状态修改无锁竞争。
func (i *impl) handleBuyActor(ctx node.Context) {
	req := &pshop.BuyRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	uid := ctx.UID()

	// 1. 查商品信息（shop 全局数据）
	items, _ := i.svc.store.ListItems(context.Background(), -1)
	var shopItem *ShopItem
	for _, it := range items {
		if it.ID == req.ShopItemID {
			shopItem = it
			break
		}
	}
	if shopItem == nil {
		stack.RespondError(ctx, stack.ErrShopItemNotFound)
		return
	}

	// 2. 扣商城库存
	if err := i.svc.store.BuyItem(context.Background(), uid, req.ShopItemID, req.Count); err != nil {
		stack.RespondError(ctx, stack.ErrShopItemSoldOut)
		return
	}

	// 3. 扣玩家金币（通过 player 服务，在 Actor 上下文安全）
	if playerSvc, ok := stack.GetService("player").(playerGoldService); ok {
		if err := playerSvc.DeductGold(uid, shopItem.Price*req.Count); err != nil {
			log.Errorf("[shop] deduct gold failed: uid=%d err=%v", uid, err)
			stack.RespondError(ctx, stack.ErrNotEnoughCurrency)
			return
		}
	}

	// 4. 同步返回结果（在 Actor goroutine 中）
	stack.RespondOK(ctx)
}

// playerGoldService 精简的玩家金币服务接口（避免循环依赖）。
type playerGoldService interface {
	DeductGold(id int64, gold int32) error
}
