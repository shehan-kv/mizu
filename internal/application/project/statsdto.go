package project

import "time"

type StatsDTO struct {
	ID             string
	Name           string
	Status         string
	CreatedAt      time.Time
	TotalTasks     int
	TasksCompleted int
	TotalInvoices  int
	InvoicesPaid   int
	TotalQuotes    int
}
