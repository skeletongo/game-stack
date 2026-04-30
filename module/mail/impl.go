package mail

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	pmail "github.com/skeletongo/game-stack/protocol/mail"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}

func (i *impl) handleList(ctx node.Context) {
	req := &pmail.ListRequest{Page: 1, PageSize: 20}
	if err := ctx.Parse(req); err != nil {
		// parse might fail if client sends empty body, use defaults
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	mails, total, err := i.svc.store.ListMail(context.Background(), ctx.UID(), req.Page, req.PageSize)
	if err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	resp := &pmail.ListResponse{
		Mails: make([]*pmail.Mail, 0, len(mails)),
		Total: total,
	}
	for _, m := range mails {
		resp.Mails = append(resp.Mails, toProto(m))
	}

	stack.RespondData(ctx, resp)
}

func (i *impl) handleRead(ctx node.Context) {
	req := &pmail.ReadRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	m, err := i.svc.store.GetMail(context.Background(), req.MailID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrNotFound)
		return
	}

	if m.ReceiverID != ctx.UID() {
		stack.RespondError(ctx, stack.ErrForbidden)
		return
	}

	if err := i.svc.store.MarkRead(context.Background(), req.MailID); err != nil {
		log.Errorf("mark mail read failed: %v", err)
	}

	stack.RespondData(ctx, &pmail.ReadResponse{Mail: toProto(m)})
}

func (i *impl) handleReceiveAttach(ctx node.Context) {
	req := &pmail.ReceiveAttachmentRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	m, err := i.svc.store.GetMail(context.Background(), req.MailID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrNotFound)
		return
	}

	if m.ReceiverID != ctx.UID() {
		stack.RespondError(ctx, stack.ErrForbidden)
		return
	}

	if err := i.svc.store.MarkAttachmentReceived(context.Background(), req.MailID, req.ItemID); err != nil {
		log.Errorf("mark attachment received failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleDelete(ctx node.Context) {
	req := &pmail.DeleteRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if len(req.MailIDs) == 0 {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if err := i.svc.store.DeleteMail(context.Background(), req.MailIDs); err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleSend(ctx node.Context) {
	req := &pmail.SendRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	attachments := make([]*Attachment, 0, len(req.Attachments))
	for _, a := range req.Attachments {
		attachments = append(attachments, &Attachment{
			ItemID:   a.ItemID,
			ItemName: a.ItemName,
			Count:    a.Count,
		})
	}

	_, err := i.svc.SendMail(ctx.UID(), "", req.ReceiverID, req.Title, req.Content, attachments, 30)
	if err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondOK(ctx)
}

func toProto(m *Mail) *pmail.Mail {
	p := &pmail.Mail{
		ID:          m.ID,
		SenderID:    m.SenderID,
		SenderName:  m.SenderName,
		ReceiverID:  m.ReceiverID,
		Title:       m.Title,
		Content:     m.Content,
		Attachments: make([]*pmail.Attachment, 0, len(m.Attachments)),
		Read:        m.Read,
		SentAt:      m.SentAt,
		ExpireAt:    m.ExpireAt,
	}
	for _, a := range m.Attachments {
		p.Attachments = append(p.Attachments, &pmail.Attachment{
			ItemID:   a.ItemID,
			ItemName: a.ItemName,
			Count:    a.Count,
			Received: a.Received,
		})
	}
	return p
}
