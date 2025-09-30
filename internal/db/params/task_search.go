package params

type TaskSearch struct {
	Keyword  string
	Status   string
	Priority string
	Offset   int64
	Limit    int64
}
