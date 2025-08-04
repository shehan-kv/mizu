package aggregates

import (
	"time"
)

type ProjectWithStats struct {
	Id             int64
	Name           string
	CreatedAt      *time.Time
	Status         string
	TotalTasks     int64
	TasksCompleted int64
	TotalInvoices  int64
	InvoicesPaid   int64
	TotalQuotes    int64
}

type ProjectWithStatsList struct {
	TotalCount int64
	Projects   []ProjectWithStats
}
