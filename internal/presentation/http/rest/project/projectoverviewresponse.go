package project

import "time"

type ProjectOverviewResponse struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	Status              string           `json:"status"`
	CreatedAt           time.Time        `json:"createdAt"`
	Members             []MemberResponse `json:"members"`
	TaskCount           int              `json:"taskCount"`
	TaskCompletedCount  int              `json:"taskCompletedCount"`
	InvoiceCount        int              `json:"invoiceCount"`
	InvoicePaidCount    int              `json:"invoicePaidCount"`
	QuoteCount          int              `json:"quoteCount"`
	ContractCount       int              `json:"contractCount"`
	ContractSignedCount int              `json:"contractSignedCount"`
	FileCount           int              `json:"fileCount"`
}
