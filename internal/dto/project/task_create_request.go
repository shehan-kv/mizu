package project

import "strings"

// Represents a project task create request
type TaskCreateRequest struct {
	Priority             string  `json:"priority"`
	Name                 string  `json:"name"`
	Status               string  `json:"status"`
	Description          string  `json:"description"`
	EstimatedTimeMinutes int64   `json:"estimatedTimeMinutes"`
	Assignees            []int64 `json:"assignees"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (r *TaskCreateRequest) Validate() bool {
	r.format()

	if len(r.Priority) == 0 {
		return false
	}

	if len(r.Name) == 0 {
		return false
	}

	if len(r.Status) == 0 {
		return false
	}

	if len(r.Description) == 0 {
		return false
	}

	if r.EstimatedTimeMinutes <= 0 {
		return false
	}

	return true
}

func (r *TaskCreateRequest) format() {
	r.Priority = strings.TrimSpace(r.Priority)
	r.Name = strings.TrimSpace(r.Name)
	r.Status = strings.TrimSpace(r.Status)
	r.Description = strings.TrimSpace(r.Description)
}
