package eventhandler

import (
	"context"
	"encoding/json"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	messageIntgEvt "mizu/internal/application/message/integration"
	"mizu/internal/presentation/http/rest/message"
	"mizu/internal/presentation/http/sse"
)

type MessageBroadcastHandler struct {
	sender *sse.Sender
	log    logger.Logger
}

func NewMessageBroadcastHandler(
	sender *sse.Sender,
	log logger.Logger,
) *MessageBroadcastHandler {
	return &MessageBroadcastHandler{
		sender: sender,
		log:    log,
	}
}

func (h *MessageBroadcastHandler) Handle(
	ctx context.Context,
	event eventbus.Event,
) error {

	msg, ok := event.(messageIntgEvt.MessageBroadcast)
	if !ok {
		h.log.Warn(
			"unexpected event type",
			"event_type", event.EventType(),
		)
		return nil
	}

	payload := message.MessageResponse{
		ID:        msg.MessageID,
		ChannelID: msg.ChannelID,
		IsSystem:  msg.IsSystem,
		Content:   msg.Content,
		CreatedAt: msg.OccurredAt,
	}

	if !payload.IsSystem {
		payload.Sender = &message.MessageSenderResponse{
			ID:        msg.SenderID,
			FirstName: *msg.SenderFirstName,
			LastName:  *msg.SenderLastName,
			Image:     msg.SenderImage,
			Title:     msg.SenderTitle,
			Role:      *msg.SenderRole,
		}
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	h.sender.SendTo(
		string(event.EventType()),
		bytes,
		msg.To,
	)

	return nil
}
