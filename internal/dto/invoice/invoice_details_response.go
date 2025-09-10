package invoice

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

type InvoiceDetailsResponse struct {
	Id           int64                    `json:"id"`
	ProjectId    int64                    `json:"projectId"`
	ProjectName  string                   `json:"projectName"`
	IsInvoice    bool                     `json:"isInvoice"`
	Status       string                   `json:"status"`
	IssuedAt     time.Time                `json:"issuedAt"`
	DueAt        *time.Time               `json:"dueAt"`
	Total        *apd.Decimal             `json:"total"`
	Discount     *apd.Decimal             `json:"discount"`
	Tax          *apd.Decimal             `json:"tax"`
	CurrencyCode string                   `json:"currencyCode"`
	Note         *string                  `json:"note"`
	Items        []InvoiceItemResponse    `json:"items"`
	History      []InvoiceHistoryResponse `json:"history"`
}
