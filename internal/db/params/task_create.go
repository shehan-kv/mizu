package params

// Parameters to create a project task
type TaskCreate struct {
	ProjectId            int64
	Priority             string
	Status               string
	Name                 string
	Description          string
	EstimatedTimeMinutes int64
	Assignees            []int64
}
