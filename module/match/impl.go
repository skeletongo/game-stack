package match

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	pmatch "github.com/skeletongo/game-stack/protocol/match"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}

func (i *impl) handleJoin(ctx node.Context) {
	req := &pmatch.JoinRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.MatchType == "" {
		req.MatchType = "pvp"
	}
	if req.Mode == "" {
		req.Mode = "1v1"
	}

	info := &MatchInfo{
		UID:       ctx.UID(),
		MatchType: req.MatchType,
		Mode:      req.Mode,
		MaxRank:   req.MaxRank,
		MinRank:   req.MinRank,
	}

	if err := i.svc.store.Push(context.Background(), ctx.UID(), info); err != nil {
		log.Errorf("match push failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondData(ctx, &pmatch.JoinResponse{
		QueueID:   "queue:" + req.MatchType + ":" + req.Mode,
		QueueSize: 1,
	})
}

func (i *impl) handleLeave(ctx node.Context) {
	if err := i.svc.store.Pop(context.Background(), ctx.UID()); err != nil {
		stack.RespondError(ctx, stack.ErrNotInQueue)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleCancel(ctx node.Context) {
	_ = i.svc.store.Pop(context.Background(), ctx.UID())
	stack.RespondOK(ctx)
}

func (i *impl) handleStatus(ctx node.Context) {
	info, err := i.svc.store.GetStatus(context.Background(), ctx.UID())
	if err != nil {
		stack.RespondData(ctx, &pmatch.StatusResponse{InQueue: false})
		return
	}

	stack.RespondData(ctx, &pmatch.StatusResponse{
		InQueue: true,
	})
	_ = info
}
