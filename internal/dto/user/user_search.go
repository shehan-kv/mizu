package user

type UserSearch struct {
	Keyword string
	Role    string
	Page    int64
	Limit   int64
}
