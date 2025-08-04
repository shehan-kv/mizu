package project

type ProjectSearchQuery struct {
	Keyword string
	Status  string
	Page    int64
	Limit   int64
}
