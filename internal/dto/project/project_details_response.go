package project

import "time"

type ProjectDetailsResponse struct {
	Id                   int64                   `json:"id"`
	Name                 string                  `json:"name"`
	CreatedAt            time.Time               `json:"createdAt"`
	Status               string                  `json:"status"`
	TaskCount            int64                   `json:"taskCount"`
	TaskCompletedCount   int64                   `json:"taskCompletedCount"`
	InvoiceCount         int64                   `json:"invoiceCount"`
	InvoicePaidCount     int64                   `json:"invoicePaidCount"`
	QuoteCount           int64                   `json:"quoteCount"`
	ContractCount        int64                   `json:"contractCount"`
	ContractSignedCount  int64                   `json:"contractSignedCount"`
	ChangeReqCount       int64                   `json:"changeReqCount"`
	ChangeReqClosedCount int64                   `json:"changeReqClosedCount"`
	FileCount            int64                   `json:"fileCount"`
	Members              []ProjectMemberResponse `json:"members"`
}
