package invoice

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type InvoiceWithProjectResponse struct {
	Id           int64        `json:"id"`
	ProjectId    int64        `json:"projectId"`
	ProjectName  string       `json:"projectName"`
	IsInvoice    bool         `json:"isInvoice"`
	IssuedAt     time.Time    `json:"issuedAt"`
	DueAt        *time.Time   `json:"dueAt"`
	Total        *apd.Decimal `json:"total"`
	CurrencyCode string       `json:"currencyCode"`
	Status       string       `json:"status"`
}
