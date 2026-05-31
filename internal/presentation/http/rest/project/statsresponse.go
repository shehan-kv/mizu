package project

import "time"

type StatsResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	TotalTasks     int       `json:"totalTasks"`
	TasksCompleted int       `json:"tasksCompleted"`
	TotalInvoices  int       `json:"totalInvoices"`
	InvoicesPaid   int       `json:"invoicesPaid"`
	TotalQuotes    int       `json:"totalQuotes"`
}
