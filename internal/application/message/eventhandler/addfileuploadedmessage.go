package eventhandler

import (
	"context"
	"encoding/json"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/message/integration"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
)

type AddFileUploadedMessage struct {
	messageRepo message.MessageRepository
	channelRepo message.ChannelRepository
	iamRepo     iam.Repository

	externalBus eventbus.ExternalBus

	idGen common.IDGenerator
}

func NewAddFileUploadedMessage(
	messageRepo message.MessageRepository,
	iamRepo iam.Repository,
	externalBus eventbus.ExternalBus,
	idGen common.IDGenerator,
) *AddFileUploadedMessage {

	return &AddFileUploadedMessage{
		messageRepo: messageRepo,
		iamRepo:     iamRepo,
		externalBus: externalBus,
		idGen:       idGen,
	}
}

func (h *AddFileUploadedMessage) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(message.FileCreatedEvent)
	if !ok {
		return nil
	}

	u, err := h.iamRepo.GetByID(ctx, e.UserID)
	if err != nil {
		return err
	}

	c, err := h.channelRepo.Get(ctx, e.ChannelID)
	if err != nil {
		return err
	}

	type payload struct {
		Type       string `json:"type"`
		FileID     string `json:"fileId"`
		Name       string `json:"name"`
		MimeType   string `json:"mimeType"`
		Size       int64  `json:"size"`
		UserID     string `json:"userId"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		UploadedAt string `json:"uploadedAt"`
	}

	p := payload{
		Type:       e.EventType().String(),
		FileID:     e.FileID.String(),
		Name:       e.OriginalName.String(),
		MimeType:   e.MimeType,
		Size:       e.Size,
		UserID:     u.ID().String(),
		FirstName:  u.FirstName(),
		LastName:   u.LastName(),
		UploadedAt: e.OccurredAt.String(),
	}

	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	content, err := message.NewContent(string(b))
	if err != nil {
		return err
	}

	id, err := h.idGen.Generate()
	if err != nil {
		return err
	}

	mID, err := message.NewMessageID(id)
	if err != nil {
		return err
	}

	msg := message.NewSystemMessage(mID, e.ChannelID, content, e.OccurredAt)

	if err := h.messageRepo.Add(ctx, msg); err != nil {
		return err
	}

	members := c.Members()
	to := make([]string, 0, len(members))
	for _, m := range members {
		to = append(to, m.String())
	}

	if err := h.externalBus.Publish(ctx, integration.MessageBroadcast{
		MessageID:  mID.String(),
		ChannelID:  e.ChannelID.String(),
		To:         to,
		SenderID:   iam.SystemUserID.String(),
		IsSystem:   true,
		Content:    content.String(),
		OccurredAt: e.OccurredAt,
	}); err != nil {
		return err
	}

	return nil
}
