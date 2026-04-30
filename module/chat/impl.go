package chat

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	pchat "github.com/skeletongo/game-stack/protocol/chat"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc  *service
	opts *options
}

func newImpl(store Store, opts *options) *impl {
	return &impl{svc: newService(store, opts), opts: opts}
}

func (i *impl) handleSend(ctx node.Context) {
	req := &pchat.SendRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if len(req.Content) > int(i.opts.maxMsgLen) {
		stack.RespondError(ctx, stack.ErrChatTooLong)
		return
	}

	uid := ctx.UID()

	msg := &Message{
		ID:       nextID(),
		SenderID: uid,
		Content:  req.Content,
		Channel:  req.Channel,
		TargetID: req.TargetID,
	}

	if err := i.svc.store.SaveMessage(context.Background(), msg); err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleHistory(ctx node.Context) {
	req := &pchat.HistoryRequest{Page: 1, PageSize: 20}
	if err := ctx.Parse(req); err != nil {
		// use defaults
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	messages, total, err := i.svc.store.GetHistory(context.Background(), req.Channel, req.TargetID, req.Page, req.PageSize)
	if err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	resp := &pchat.HistoryResponse{
		Messages: make([]*pchat.ChatMessage, 0, len(messages)),
		Total:    total,
	}
	for _, m := range messages {
		resp.Messages = append(resp.Messages, &pchat.ChatMessage{
			ID:         m.ID,
			SenderID:   m.SenderID,
			SenderName: m.SenderName,
			Content:    m.Content,
			Channel:    m.Channel,
			TargetID:   m.TargetID,
			SentAt:     m.SentAt,
		})
	}

	stack.RespondData(ctx, resp)
}

func (i *impl) handleWorldChat(ctx node.Context) {
	req := &pchat.SendRequest{Channel: "world"}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.Channel == "" {
		req.Channel = "world"
	}

	ctx.Task(func(ctx node.Context) {
		i.handleSend(ctx)
	})
	log.Debugf("world chat message: uid=%d content=%s", ctx.UID(), req.Content)
}

func (i *impl) handlePrivateChat(ctx node.Context) {
	req := &pchat.SendRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.TargetID == 0 {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	req.Channel = "private"

	ctx.Task(func(ctx node.Context) {
		i.handleSend(ctx)
	})
}

func (i *impl) handleGuildChat(ctx node.Context) {
	req := &pchat.SendRequest{Channel: "guild"}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	req.Channel = "guild"

	ctx.Task(func(ctx node.Context) {
		i.handleSend(ctx)
	})
}
