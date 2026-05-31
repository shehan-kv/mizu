package eventhandler

import (
	"context"
	"mizu/internal/application/uow"
	"mizu/internal/domain/common"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

type CreateProjectChannel struct {
	channelRepo message.ChannelRepository

	idGen common.IDGenerator
	uow   uow.UnitOfWork
}

func NewCreateProjectChannel(
	channelRepo message.ChannelRepository,
	idgen common.IDGenerator,
	uow uow.UnitOfWork) *CreateProjectChannel {

	return &CreateProjectChannel{
		channelRepo: channelRepo,
		idGen:       idgen,
		uow:         uow,
	}
}

func (h *CreateProjectChannel) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(project.ProjectCreatedEvent)
	if !ok {
		return nil
	}

	id, err := h.idGen.Generate()
	if err != nil {
		return err
	}

	cID, err := message.NewChannelID(id)
	if err != nil {
		return err
	}

	name, err := message.NewChannelName(e.Name.String())
	if err != nil {
		return err
	}

	c, err := message.NewChannel(cID, &e.ProjectID, name, e.Members, e.OccurredAt)
	if err != nil {
		return err
	}

	return h.uow.Execute(ctx, func(ctx context.Context) error {
		return h.channelRepo.Add(ctx, c)
	})
}
