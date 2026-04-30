package inventory

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"

	pinv "github.com/skeletongo/game-stack/protocol/inventory"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc  *service
	opts *options
}

func newImpl(store Store, opts *options) *impl {
	return &impl{svc: newService(store), opts: opts}
}

func (i *impl) handleList(ctx node.Context) {
	req := &pinv.ListRequest{}
	_ = ctx.Parse(req)

	items, err := i.svc.store.ListItems(context.Background(), ctx.UID(), req.BagType)
	if err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	resp := &pinv.ListResponse{
		Items:     make([]*pinv.Item, 0, len(items)),
		BagSlots:  i.opts.bagSlots,
		UsedSlots: int32(len(items)),
	}
	for _, it := range items {
		resp.Items = append(resp.Items, toProtoItem(it))
	}

	stack.RespondData(ctx, resp)
}

func (i *impl) handleUse(ctx node.Context) {
	req := &pinv.UseRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.UseItem(context.Background(), ctx.UID(), req.ID, req.Count); err != nil {
		stack.RespondError(ctx, stack.ErrCannotUse)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleEquip(ctx node.Context) {
	req := &pinv.EquipRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.EquipItem(context.Background(), ctx.UID(), req.ID); err != nil {
		stack.RespondError(ctx, stack.ErrCannotEquip)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleUnequip(ctx node.Context) {
	req := &pinv.UnequipRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.UnequipItem(context.Background(), ctx.UID(), req.ID); err != nil {
		stack.RespondError(ctx, stack.ErrCannotEquip)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleDrop(ctx node.Context) {
	req := &pinv.DropRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.RemoveItem(context.Background(), ctx.UID(), req.ID, req.Count); err != nil {
		stack.RespondError(ctx, stack.ErrItemNotEnough)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleSell(ctx node.Context) {
	req := &pinv.SellRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.RemoveItem(context.Background(), ctx.UID(), req.ID, req.Count); err != nil {
		stack.RespondError(ctx, stack.ErrItemNotEnough)
		return
	}

	// 简单计算售价（1金币/件）
	stack.RespondData(ctx, &pinv.SellResponse{Gold: req.Count})
}

func toProtoItem(it *Item) *pinv.Item {
	return &pinv.Item{
		ID:       it.ID,
		ItemID:   it.ItemID,
		Name:     it.Name,
		Type:     it.Type,
		Count:    it.Count,
		MaxStack: it.MaxStack,
		Level:    it.Level,
		Quality:  it.Quality,
		Equipped: it.Equipped,
	}
}
