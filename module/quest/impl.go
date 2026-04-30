package quest

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"

	pquest "github.com/skeletongo/game-stack/protocol/quest"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}

func (i *impl) handleList(ctx node.Context) {
	req := &pquest.ListRequest{}
	_ = ctx.Parse(req)

	quests, err := i.svc.store.ListPlayerQuests(context.Background(), ctx.UID(), req.Type)
	if err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	resp := &pquest.ListResponse{Quests: make([]*pquest.Quest, 0, len(quests))}
	for _, q := range quests {
		resp.Quests = append(resp.Quests, &pquest.Quest{
			ID:          q.ID,
			Name:        q.Name,
			Description: q.Description,
			Type:        q.Type,
			Status:      int32(q.Status),
			Progress:    q.Progress,
			Target:      q.Target,
			RewardGold:  q.RewardGold,
			RewardExp:   q.RewardExp,
			RewardItems: q.RewardItems,
		})
	}

	stack.RespondData(ctx, resp)
}

func (i *impl) handleAccept(ctx node.Context) {
	req := &pquest.AcceptRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.AcceptQuest(context.Background(), ctx.UID(), req.QuestID); err != nil {
		stack.RespondError(ctx, stack.ErrQuestNotFound)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleSubmit(ctx node.Context) {
	req := &pquest.SubmitRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	q, err := i.svc.store.GetPlayerQuest(context.Background(), ctx.UID(), req.QuestID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrQuestNotFound)
		return
	}

	if q.Status != QuestDoing {
		stack.RespondError(ctx, stack.ErrQuestNotAccepted)
		return
	}
	if q.Progress < q.Target {
		stack.RespondError(ctx, stack.ErrQuestNotComplete)
		return
	}

	if err := i.svc.store.SubmitQuest(context.Background(), ctx.UID(), req.QuestID); err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondData(ctx, &pquest.SubmitResponse{
		RewardGold:  q.RewardGold,
		RewardExp:   q.RewardExp,
		RewardItems: q.RewardItems,
	})
}

func (i *impl) handleAbandon(ctx node.Context) {
	req := &pquest.AbandonRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	_ = i.svc.store.UpdateProgress(context.Background(), ctx.UID(), req.QuestID, 0)
	stack.RespondOK(ctx)
}
