package task

import "time"

type TaskResponse struct {
	ID               string             `json:"id"`
	ProjectID        string             `json:"projectId"`
	Priority         string             `json:"priority"`
	Status           string             `json:"status"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	EstimatedMinutes int                `json:"estimatedMinutes"`
	Assignees        []AssigneeResponse `json:"assignees"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}
