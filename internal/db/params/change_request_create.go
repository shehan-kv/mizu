package params

type ChangeRequestCreate struct {
	ProjectId int64
	Status    string
	Title     string
	Content   string
}
