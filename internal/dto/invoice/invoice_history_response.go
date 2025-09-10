package invoice

import "time"

type InvoiceHistoryResponse struct {
	Id         int64              `json:"id"`
	User       InvoiceHistoryUser `json:"user"`
	Event      string             `json:"event"`
	RecoredAt  time.Time          `json:"recoredAt"`
	IsInvoice  bool               `json:"isInvoice"`
	LastStatus *string            `json:"lastStatus"`
	NewStatus  *string            `json:"newStatus"`
}
