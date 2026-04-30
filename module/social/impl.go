package social

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	psocial "github.com/skeletongo/game-stack/protocol/social"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct{ svc *service }

func newImpl(store Store) *impl { return &impl{svc: newService(store)} }

func (i *impl) handleFriendList(ctx node.Context) {
	req := &psocial.FriendListRequest{Page: 1, PageSize: 50}
	_ = ctx.Parse(req)
	friends, total, max, _ := i.svc.store.ListFriends(context.Background(), ctx.UID(), req.Page, req.PageSize)
	resp := &psocial.FriendListResponse{Friends: make([]*psocial.FriendInfo, 0, len(friends)), Total: total, MaxCount: max}
	for _, f := range friends {
		resp.Friends = append(resp.Friends, &psocial.FriendInfo{PlayerID: f.PlayerID, Nickname: f.Nickname, Level: f.Level, Avatar: f.Avatar, Online: f.Online, GuildName: f.GuildName, Intimacy: f.Intimacy})
	}
	stack.RespondData(ctx, resp)
}

func (i *impl) handleAdd(ctx node.Context) {
	req := &psocial.AddFriendRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	if req.PlayerID == ctx.UID() {
		stack.RespondError(ctx, stack.ErrFriendSelf)
		return
	}
	f := &FriendInfo{PlayerID: req.PlayerID}
	if err := i.svc.store.AddFriend(context.Background(), ctx.UID(), f); err != nil {
		stack.RespondError(ctx, stack.ErrFriendAlready)
		return
	}
	stack.RespondOK(ctx)
}

func (i *impl) handleRemove(ctx node.Context) {
	req := &psocial.RemoveFriendRequest{}
	_ = ctx.Parse(req)
	_ = i.svc.store.RemoveFriend(context.Background(), ctx.UID(), req.PlayerID)
	stack.RespondOK(ctx)
}

func (i *impl) handleBlock(ctx node.Context) {
	req := &psocial.BlockRequest{}
	_ = ctx.Parse(req)
	_ = i.svc.store.BlockUser(context.Background(), ctx.UID(), req.PlayerID)
	stack.RespondOK(ctx)
}

func (i *impl) handleUnblock(ctx node.Context) {
	req := &psocial.UnblockRequest{}
	_ = ctx.Parse(req)
	_ = i.svc.store.UnblockUser(context.Background(), ctx.UID(), req.PlayerID)
	stack.RespondOK(ctx)
}

func (i *impl) handleBlacklist(ctx node.Context) {
	req := &psocial.BlacklistRequest{Page: 1, PageSize: 50}
	_ = ctx.Parse(req)
	list, total, _ := i.svc.store.ListBlacklist(context.Background(), ctx.UID(), req.Page, req.PageSize)
	resp := &psocial.BlacklistResponse{Blacklist: make([]*psocial.FriendInfo, 0, len(list)), Total: total}
	for _, f := range list {
		resp.Blacklist = append(resp.Blacklist, &psocial.FriendInfo{PlayerID: f.PlayerID, Nickname: f.Nickname, Level: f.Level})
	}
	stack.RespondData(ctx, resp)
}
