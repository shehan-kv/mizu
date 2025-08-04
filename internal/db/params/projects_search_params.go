package params

type ProjectsSearchParams struct {
	Keyword string
	Status  string
	Offset  int64
	Limit   int64
	UserId  int64
}
