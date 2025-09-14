package params

type ChangeRequestSearch struct {
	Keyword string
	Status  string
	Offset  int64
	Limit   int64
}
