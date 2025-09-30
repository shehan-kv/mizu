package project

// Represents a query on project tasks
type TaskSearchQuery struct {
	Keyword  string
	Status   string
	Priority string
	Page     int64
	Limit    int64
}
