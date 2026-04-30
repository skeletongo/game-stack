package room

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	proom "github.com/skeletongo/game-stack/protocol/room"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc  *service
	opts *options
}

func newImpl(store Store, opts *options) *impl {
	return &impl{svc: newService(store), opts: opts}
}

func (i *impl) handleCreate(ctx node.Context) {
	req := &proom.CreateRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.MaxPlayers <= 0 {
		req.MaxPlayers = i.opts.maxPlayers
	}
	if req.MaxPlayers > i.opts.maxPlayers {
		req.MaxPlayers = i.opts.maxPlayers
	}

	room := &Room{
		ID:         time.Now().UnixNano(),
		Name:       req.Name,
		OwnerID:    ctx.UID(),
		MaxPlayers: req.MaxPlayers,
		CurPlayers: 1,
		SceneID:    req.SceneID,
		Password:   req.Password,
		Players: []*RoomPlayer{{
			PlayerID: ctx.UID(),
			JoinedAt: time.Now().Unix(),
		}},
		CreatedAt: time.Now().Unix(),
	}

	if err := i.svc.store.CreateRoom(context.Background(), room); err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondData(ctx, &proom.CreateResponse{
		Room: toProtoRoom(room),
	})
}

func (i *impl) handleJoin(ctx node.Context) {
	req := &proom.JoinRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	r, err := i.svc.store.GetRoom(context.Background(), req.RoomID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrRoomNotFound)
		return
	}

	if r.CurPlayers >= r.MaxPlayers {
		stack.RespondError(ctx, stack.ErrRoomFull)
		return
	}

	if r.Password != "" && r.Password != req.Password {
		stack.RespondError(ctx, stack.ErrRoomLocked)
		return
	}

	player := &RoomPlayer{
		PlayerID: ctx.UID(),
		JoinedAt: time.Now().Unix(),
	}

	if err := i.svc.store.AddPlayer(context.Background(), req.RoomID, player); err != nil {
		log.Errorf("add room player failed: %v", err)
		stack.RespondError(ctx, stack.ErrRoomAlreadyIn)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleLeave(ctx node.Context) {
	req := &proom.LeaveRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.RemovePlayer(context.Background(), req.RoomID, ctx.UID()); err != nil {
		stack.RespondError(ctx, stack.ErrRoomNotFound)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleList(ctx node.Context) {
	req := &proom.ListRequest{Page: 1, PageSize: 20}
	if err := ctx.Parse(req); err != nil {
		// use defaults
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	rooms, total, err := i.svc.store.ListRooms(context.Background(), req.SceneID, req.Page, req.PageSize)
	if err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	resp := &proom.ListResponse{
		Rooms: make([]*proom.RoomInfo, 0, len(rooms)),
		Total: total,
	}
	for _, r := range rooms {
		resp.Rooms = append(resp.Rooms, toProtoRoom(r))
	}

	stack.RespondData(ctx, resp)
}

func (i *impl) handleKick(ctx node.Context) {
	req := &proom.KickRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	r, err := i.svc.store.GetRoom(context.Background(), req.RoomID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrRoomNotFound)
		return
	}

	if r.OwnerID != ctx.UID() {
		stack.RespondError(ctx, stack.ErrRoomNotOwner)
		return
	}

	if err := i.svc.store.RemovePlayer(context.Background(), req.RoomID, req.PlayerID); err != nil {
		stack.RespondError(ctx, stack.ErrRoomNotFound)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleReady(ctx node.Context) {
	req := &proom.LeaveRequest{} // reuse roomID field
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.SetReady(context.Background(), req.RoomID, ctx.UID()); err != nil {
		stack.RespondError(ctx, stack.ErrRoomNotFound)
		return
	}

	stack.RespondOK(ctx)
}

func toProtoRoom(r *Room) *proom.RoomInfo {
	return &proom.RoomInfo{
		ID:         r.ID,
		Name:       r.Name,
		OwnerID:    r.OwnerID,
		MaxPlayers: r.MaxPlayers,
		CurPlayers: r.CurPlayers,
		SceneID:    r.SceneID,
		Locked:     r.Locked,
	}
}
