package eventhandler

import (
	"context"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
)

type CreateDefaultChannel struct {
	channelRepo message.ChannelRepository

	idGen common.IDGenerator
}

func NewCreateDefaultChannel(
	mailer mailer.Mailer,
	channelRepo message.ChannelRepository,
	idGen common.IDGenerator,
) *CreateDefaultChannel {

	return &CreateDefaultChannel{channelRepo: channelRepo, idGen: idGen}
}

func (h *CreateDefaultChannel) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(iam.UserCreatedEvent)
	if !ok {
		return nil
	}

	id, err := h.idGen.Generate()
	if err != nil {
		return err
	}

	chID, err := message.NewChannelID(id)
	if err != nil {
		return err
	}

	chName, err := message.NewChannelName(e.Name.FirstName() + " " + e.Name.LastName() + " · Workspace")
	if err != nil {
		return err
	}

	ch, err := message.NewChannel(chID, nil, chName, []iam.UserID{e.UserID, e.CreatedBy}, e.OccurredAt)
	if err != nil {
		return err
	}

	return h.channelRepo.Save(ctx, ch)
}
