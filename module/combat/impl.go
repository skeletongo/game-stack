package combat

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"

	pcombat "github.com/skeletongo/game-stack/protocol/combat"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}

func (i *impl) handleSkillCast(ctx node.Context) {
	req := &pcombat.CastRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	damage, crit, err := i.svc.CalcDamage(ctx.UID(), req.TargetID, req.SkillID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrSkillNotFound)
		return
	}

	_, _ = i.svc.store.UpdateHP(context.Background(), req.TargetID, -damage)

	stack.RespondData(ctx, &pcombat.CastResponse{
		SkillID:  req.SkillID,
		CasterID: ctx.UID(),
		TargetID: req.TargetID,
		Damage:   damage,
		Critical: crit,
	})
}

func (i *impl) handleMove(ctx node.Context) {
	req := &pcombat.MoveRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	_ = req // position sync
	stack.RespondOK(ctx)
}

func (i *impl) handleTarget(ctx node.Context) {
	req := &pcombat.CastRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	// validate target
	_ = req
	stack.RespondOK(ctx)
}
