package activity

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	pact "github.com/skeletongo/game-stack/protocol/activity"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct{ svc *service }

func newImpl(store Store) *impl { return &impl{svc: newService(store)} }

func (i *impl) handleList(ctx node.Context) {
	req := &pact.ListRequest{Type: -1}
	_ = ctx.Parse(req)
	activities, _ := i.svc.store.ListActivities(context.Background(), req.Type)
	resp := &pact.ListResponse{Activities: make([]*pact.ActivityInfo, 0, len(activities))}
	for _, a := range activities {
		resp.Activities = append(resp.Activities, &pact.ActivityInfo{
			ID: a.ID, Name: a.Name, Description: a.Description,
			Type: a.Type, Status: a.Status, StartAt: a.StartAt, EndAt: a.EndAt,
			Progress: a.Progress, Target: a.Target, Claimed: a.Claimed, Rewards: a.Rewards,
		})
	}
	stack.RespondData(ctx, resp)
}

func (i *impl) handleClaim(ctx node.Context) {
	req := &pact.ClaimRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	if err := i.svc.store.ClaimReward(context.Background(), ctx.UID(), req.ActivityID); err != nil {
		stack.RespondError(ctx, stack.ErrActivityClaimed)
		return
	}
	stack.RespondOK(ctx)
}

func (i *impl) handleInfo(ctx node.Context) {
	req := &pact.ListRequest{Type: -1}
	_ = ctx.Parse(req)
	a, err := i.svc.store.GetActivity(context.Background(), req.Type)
	if err != nil {
		stack.RespondError(ctx, stack.ErrActivityNotFound)
		return
	}
	stack.RespondData(ctx, &pact.ActivityInfo{
		ID: a.ID, Name: a.Name, Description: a.Description,
		Type: a.Type, Status: a.Status, StartAt: a.StartAt, EndAt: a.EndAt,
		Progress: a.Progress, Target: a.Target, Claimed: a.Claimed, Rewards: a.Rewards,
	})
}
