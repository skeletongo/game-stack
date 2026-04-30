package guild

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	pguild "github.com/skeletongo/game-stack/protocol/guild"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct{ svc *service }

func newImpl(store Store) *impl { return &impl{svc: newService(store)} }

func (i *impl) handleCreate(ctx node.Context) {
	req := &pguild.CreateRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	g := &Guild{
		ID: time.Now().UnixNano(), Name: req.Name, Level: 1,
		OwnerID: ctx.UID(), MemberCount: 1, MaxMembers: 50,
		Members:   []*Member{{PlayerID: ctx.UID(), Position: 3, JoinedAt: time.Now().Unix()}},
		CreatedAt: time.Now().Unix(),
	}
	if err := i.svc.store.CreateGuild(context.Background(), g); err != nil {
		stack.RespondError(ctx, stack.ErrGuildNameExists)
		return
	}
	stack.RespondData(ctx, &pguild.CreateResponse{Guild: toProtoGuild(g)})
}

func (i *impl) handleJoin(ctx node.Context) {
	req := &pguild.JoinRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	m := &Member{PlayerID: ctx.UID(), Position: 0, JoinedAt: time.Now().Unix()}
	if err := i.svc.store.AddMember(context.Background(), req.GuildID, m); err != nil {
		stack.RespondError(ctx, stack.ErrGuildFull)
		return
	}
	stack.RespondOK(ctx)
}

func (i *impl) handleLeave(ctx node.Context) {
	req := &pguild.LeaveRequest{}
	_ = ctx.Parse(req)
	_ = i.svc.store.RemoveMember(context.Background(), req.GuildID, ctx.UID())
	stack.RespondOK(ctx)
}

func (i *impl) handleKick(ctx node.Context) {
	req := &pguild.KickRequest{}
	_ = ctx.Parse(req)
	g, _ := i.svc.store.GetGuild(context.Background(), req.GuildID)
	if g == nil || g.OwnerID != ctx.UID() {
		stack.RespondError(ctx, stack.ErrGuildInsufficientRank)
		return
	}
	_ = i.svc.store.RemoveMember(context.Background(), req.GuildID, req.PlayerID)
	stack.RespondOK(ctx)
}

func (i *impl) handleList(ctx node.Context) {
	req := &pguild.ListRequest{Page: 1, PageSize: 20}
	_ = ctx.Parse(req)
	guilds, total, _ := i.svc.store.ListGuilds(context.Background(), req.Page, req.PageSize)
	resp := &pguild.ListResponse{Guilds: make([]*pguild.GuildInfo, 0, len(guilds)), Total: total}
	for _, g := range guilds {
		resp.Guilds = append(resp.Guilds, toProtoGuild(g))
	}
	stack.RespondData(ctx, resp)
}

func (i *impl) handleInfo(ctx node.Context) {
	req := &pguild.InfoRequest{}
	_ = ctx.Parse(req)
	g, err := i.svc.store.GetGuild(context.Background(), req.GuildID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrGuildNotFound)
		return
	}
	members := make([]*pguild.Member, 0, len(g.Members))
	for _, m := range g.Members {
		members = append(members, &pguild.Member{PlayerID: m.PlayerID, Nickname: m.Nickname, Level: m.Level, Position: m.Position, Donate: m.Donate, JoinedAt: m.JoinedAt})
	}
	stack.RespondData(ctx, &pguild.InfoResponse{Guild: toProtoGuild(g), Members: members})
}

func (i *impl) handleDonate(ctx node.Context) {
	req := &pguild.DonateRequest{}
	_ = ctx.Parse(req)
	exp, err := i.svc.store.Donate(context.Background(), req.GuildID, ctx.UID(), req.Gold)
	if err != nil {
		stack.RespondError(ctx, stack.ErrNotEnoughCurrency)
		return
	}
	stack.RespondData(ctx, &pguild.DonateResponse{GuildExp: exp, Contribute: int64(req.Gold)})
}

func toProtoGuild(g *Guild) *pguild.GuildInfo {
	return &pguild.GuildInfo{
		ID: g.ID, Name: g.Name, Level: g.Level, Exp: g.Exp,
		OwnerID: g.OwnerID, OwnerName: g.OwnerName,
		MemberCount: g.MemberCount, MaxMembers: g.MaxMembers,
		Notice: g.Notice, Gold: g.Gold, CreatedAt: g.CreatedAt,
	}
}
