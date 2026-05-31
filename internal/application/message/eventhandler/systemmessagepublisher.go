package eventhandler

import (
	"context"
	"errors"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/message/integration"
	"mizu/internal/application/uow"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type SystemMessagePublisher struct {
	channelRepo message.ChannelRepository
	messageRepo message.MessageRepository

	externalBus eventbus.ExternalBus

	uow   uow.UnitOfWork
	idGen common.IDGenerator
}

// NewSystemMessagePublisher creates a new publisher service.
func NewSystemMessagePublisher(
	channelRepo message.ChannelRepository,
	messageRepo message.MessageRepository,
	externalBus eventbus.ExternalBus,
	uow uow.UnitOfWork,
	idGen common.IDGenerator,
) *SystemMessagePublisher {
	return &SystemMessagePublisher{
		channelRepo: channelRepo,
		messageRepo: messageRepo,
		externalBus: externalBus,
		uow:         uow,
		idGen:       idGen,
	}
}

// CreateAndPublish creates system messages for all channels in a project,
// persists them, and broadcasts events.
func (p *SystemMessagePublisher) CreateAndPublish(
	ctx context.Context,
	projectID project.ProjectID,
	content message.Content,
	occurredAt time.Time,
) error {

	chs, err := p.channelRepo.ListByProject(ctx, projectID)
	if err != nil {
		return err
	}

	if len(chs) == 0 {
		return nil
	}

	msgs := make([]*message.Message, 0, len(chs))
	events := make([]integration.MessageBroadcast, 0, len(chs))

	for _, ch := range chs {
		id, err := p.idGen.Generate()
		if err != nil {
			return err
		}

		mID, err := message.NewMessageID(id)
		if err != nil {
			return err
		}

		msgs = append(msgs,
			message.NewSystemMessage(
				mID,
				ch.ID(),
				content,
				occurredAt,
			),
		)

		members := ch.Members()
		to := make([]string, 0, len(members))
		for _, m := range members {
			to = append(to, m.String())
		}

		events = append(events, integration.MessageBroadcast{
			MessageID:  mID.String(),
			ChannelID:  ch.ID().String(),
			To:         to,
			SenderID:   iam.SystemUserID.String(),
			IsSystem:   true,
			Content:    content.String(),
			OccurredAt: occurredAt,
		})
	}

	// Persist messages in a transaction
	if err := p.uow.Execute(ctx, func(ctx context.Context) error {
		for _, msg := range msgs {
			if err := p.messageRepo.Add(ctx, msg); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return p.publish(ctx, events)
}

// publish sends broadcast events concurrently and collects errors.
func (p *SystemMessagePublisher) publish(
	ctx context.Context,
	events []integration.MessageBroadcast,
) error {

	g := new(errgroup.Group)

	var (
		mu   sync.Mutex
		errs []error
	)

	for _, evt := range events {
		evt := evt

		g.Go(func() error {
			if err := p.externalBus.Publish(ctx, evt); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			return nil
		})
	}

	_ = g.Wait()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
