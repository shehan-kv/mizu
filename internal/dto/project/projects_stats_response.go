package project

import (
	"time"
)

// Represents a project response with stats
type ProjectsStatsResponse struct {
	Id             int64      `json:"id"`
	Name           string     `json:"name"`
	CreatedAt      *time.Time `json:"createdAt"`
	Status         string     `json:"status"`
	TotalTasks     int64      `json:"totalTasks"`
	TasksCompleted int64      `json:"tasksCompleted"`
	TotalInvoices  int64      `json:"totalInvoices"`
	InvoicesPaid   int64      `json:"invoicesPaid"`
	TotalQuotes    int64      `json:"totalQuotes"`
}
