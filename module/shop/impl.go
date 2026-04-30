package shop

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	pshop "github.com/skeletongo/game-stack/protocol/shop"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct{ svc *service }

func newImpl(store Store) *impl { return &impl{svc: newService(store)} }

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

func (i *impl) handleBuy(ctx node.Context) {
	req := &pshop.BuyRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	if err := i.svc.store.BuyItem(context.Background(), ctx.UID(), req.ShopItemID, req.Count); err != nil {
		stack.RespondError(ctx, stack.ErrNotEnoughCurrency)
		return
	}
	stack.RespondOK(ctx)
}
