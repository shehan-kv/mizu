package eventhandler

import (
	"context"
	"encoding/json"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/message"
)

type AddInvoiceConvertedMessage struct {
	publisher *SystemMessagePublisher
}

func NewAddInvoiceConvertedMessage(publisher *SystemMessagePublisher) *AddInvoiceConvertedMessage {

	return &AddInvoiceConvertedMessage{
		publisher: publisher,
	}
}

func (h *AddInvoiceConvertedMessage) Handle(ctx context.Context, event common.Event) error {
	e, ok := event.(billing.InvoiceConvertedEvent)
	if !ok {
		return nil
	}

	type payload struct {
		Type         string `json:"type"`
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
