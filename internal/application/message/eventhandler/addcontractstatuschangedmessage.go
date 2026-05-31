package eventhandler

import (
	"context"
	"encoding/json"
	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"time"
)

type AddContractStatusChangedMessage struct {
	iamRepo iam.Repository

	publisher *SystemMessagePublisher
}

func NewAddContractStatusChangedMessage(
	iamRepo iam.Repository,
	publisher *SystemMessagePublisher,
) *AddContractStatusChangedMessage {

	return &AddContractStatusChangedMessage{
		iamRepo:   iamRepo,
		publisher: publisher,
	}
}

func (h *AddContractStatusChangedMessage) Handle(ctx context.Context, event common.Event) error {
	e, ok := event.(contract.ContractStatusChangedEvent)
	if !ok {
		return nil
	}

	u, err := h.iamRepo.GetByID(ctx, e.Signatory.UserID())
	if err != nil {
		return err
	}

	type payload struct {
		Type       string `json:"type"`
		Status     string `json:"status"`
		ContractID string `json:"contractId"`
		Name       string `json:"name"`
		UserID     string `json:"userId"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		OccurredAt string `json:"occurredAt"`
	}

	p := payload{
		Type:       e.EventType().String(),
		Status:     e.Status.String(),
		ContractID: e.ContractID.String(),
		Name:       e.Name.String(),
		UserID:     u.ID().String(),
		FirstName:  u.FirstName(),
		LastName:   u.LastName(),
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
