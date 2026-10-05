package task

type CreateTaskParams struct {
	ActorID          string
	ProjectID        string
	Priority         string
	Status           string
	Name             string
	Description      string
	EstimatedMinutes int
	AssigneeIDs      []string
}
