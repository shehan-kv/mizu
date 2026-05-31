package eventhandler

import (
	"context"
	"encoding/json"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/message"
)

type AddInvoiceStatusChangedMessage struct {
	publisher *SystemMessagePublisher
}

func NewAddInvoiceStatusChangedMessage(publisher *SystemMessagePublisher) *AddInvoiceStatusChangedMessage {

	return &AddInvoiceStatusChangedMessage{
		publisher: publisher,
	}
}

func (h *AddInvoiceStatusChangedMessage) Handle(ctx context.Context, event common.Event) error {

	e, ok := event.(billing.InvoiceStatusChangedEvent)
	if !ok {
		return nil
	}

	type payload struct {
		Type         string `json:"type"`
		Status       string `json:"status"`
		DocumentType string `json:"documentType"`
		InvoiceID    string `json:"invoiceId"`
		CurrencyCode string `json:"currencyCode"`
		SubTotal     string `json:"subTotal"`
	}

	docType := "quote"
	if e.IsInvoice {
		docType = "invoice"
	}

	p := payload{
		Type:         e.EventType().String(),
		Status:       e.Status.String(),
		DocumentType: docType,
		InvoiceID:    e.InvoiceID.String(),
		CurrencyCode: e.CurrencyCode.String(),
		SubTotal:     e.SubTotal.String(),
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
