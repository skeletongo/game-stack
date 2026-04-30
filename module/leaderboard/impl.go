package leaderboard

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	plb "github.com/skeletongo/game-stack/protocol/leaderboard"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct{ svc *service }

func newImpl(store Store) *impl { return &impl{svc: newService(store)} }

func (i *impl) handleGet(ctx node.Context) {
	req := &plb.GetRequest{BoardName: "level", Page: 1, PageSize: 50}
	_ = ctx.Parse(req)
	entries, total, _ := i.svc.store.GetTop(context.Background(), req.BoardName, req.Page, req.PageSize)
	resp := &plb.GetResponse{Entries: make([]*plb.Entry, 0, len(entries)), BoardName: req.BoardName, Total: total}
	for _, e := range entries {
		resp.Entries = append(resp.Entries, &plb.Entry{Rank: e.Rank, PlayerID: e.PlayerID, Nickname: e.Nickname, Score: e.Score, Level: e.Level})
	}
	stack.RespondData(ctx, resp)
}

func (i *impl) handleRank(ctx node.Context) {
	req := &plb.RankRequest{BoardName: "level"}
	_ = ctx.Parse(req)
	entry, err := i.svc.store.GetRank(context.Background(), req.BoardName, ctx.UID())
	if err != nil {
		stack.RespondError(ctx, stack.ErrRankNotSet)
		return
	}
	stack.RespondData(ctx, &plb.RankResponse{Entry: &plb.Entry{Rank: entry.Rank, PlayerID: entry.PlayerID, Nickname: entry.Nickname, Score: entry.Score, Level: entry.Level}})
}

func (i *impl) handleTop(ctx node.Context) {
	req := &plb.GetRequest{BoardName: "level", Page: 1, PageSize: 10}
	_ = ctx.Parse(req)
	entries, total, _ := i.svc.store.GetTop(context.Background(), req.BoardName, req.Page, req.PageSize)
	resp := &plb.GetResponse{Entries: make([]*plb.Entry, 0, len(entries)), BoardName: req.BoardName, Total: total}
	for _, e := range entries {
		resp.Entries = append(resp.Entries, &plb.Entry{Rank: e.Rank, PlayerID: e.PlayerID, Nickname: e.Nickname, Score: e.Score, Level: e.Level})
	}
	stack.RespondData(ctx, resp)
}
