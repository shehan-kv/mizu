package project

import "time"

type ProjectOverviewDTO struct {
	ID                  string
	Name                string
	Status              string
	CreatedAt           time.Time
	Members             []MemberDTO
	TaskCount           int
	TaskCompletedCount  int
	InvoiceCount        int
	InvoicePaidCount    int
	QuoteCount          int
	ContractCount       int
	ContractSignedCount int
	FileCount           int
}
