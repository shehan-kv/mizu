package task

type CreateTaskRequest struct {
	Priority         string   `json:"priority"`
	Status           string   `json:"status"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	EstimatedMinutes int      `json:"estimatedMinutes"`
	AssigneeIDs      []string `json:"assigneeIds"`
}
