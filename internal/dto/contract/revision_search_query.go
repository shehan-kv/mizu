package contract

// Represents a query on contract revisions
type RevisionSearchQuery struct {
	Keyword string
	Status  string
	Page    int64
	Limit   int64
}
