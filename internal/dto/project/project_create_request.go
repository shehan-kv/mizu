package project

import "strings"

// Represents a project create request
type ProjectCreateRequest struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Members []int64 `json:"members"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (r *ProjectCreateRequest) Validate() bool {
	r.format()

	if len(r.Name) == 0 {
		return false
	}

	if len(r.Status) == 0 {
		return false
	}

	return true
}

func (r *ProjectCreateRequest) format() {
	r.Name = strings.TrimSpace(r.Name)
	r.Status = strings.TrimSpace(r.Status)
}
