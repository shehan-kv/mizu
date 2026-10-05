package task

import "time"

type TaskDTO struct {
	ID               string
	ProjectID        string
	Priority         string
	Status           string
	Name             string
	Description      string
	EstimatedMinutes int
	Assignees        []AssigneeDTO

	CreatedAt time.Time
	UpdatedAt time.Time
}
