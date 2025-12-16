package project

type MemberSearchQuery struct {
	Keyword string
	Role    string
	Page    int64
	Limit   int64
}
