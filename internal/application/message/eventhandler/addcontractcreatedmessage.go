package eventhandler

import (
	"context"
	"encoding/json"
	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/message"
	"time"
)

type AddContractCreatedMessage struct {
	publisher *SystemMessagePublisher
}

func NewAddContractCreatedMessage(publisher *SystemMessagePublisher) *AddContractCreatedMessage {

	return &AddContractCreatedMessage{
		publisher: publisher,
	}
}

func (h *AddContractCreatedMessage) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(contract.ContractCreatedEvent)
	if !ok {
		return nil
	}

	type payload struct {
		Type       string `json:"type"`
		ContractID string `json:"contract_id"`
		Name       string `json:"name"`
		OccurredAt string `json:"occurred_at"`
	}

	p := payload{
		Type:       e.EventType().String(),
		ContractID: e.ContractID.String(),
		Name:       e.Name.String(),
		OccurredAt: e.OccurredAt.Format(time.RFC3339Nano),
	}

	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	content, err := message.NewContent(string(b))
	if err != nil {
		return err
	}

	return h.publisher.CreateAndPublish(ctx, e.ProjectID, content, e.OccurredAt)
}
