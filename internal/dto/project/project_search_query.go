package project

// Represents a query on projects
type ProjectSearchQuery struct {
	Keyword string
	Status  string
	Page    int64
	Limit   int64
}
