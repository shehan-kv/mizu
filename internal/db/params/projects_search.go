package params

type ProjectsSearch struct {
	Keyword string
	Status  string
	Offset  int64
	Limit   int64
	UserId  int64
}
