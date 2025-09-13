package changerequest

import "strings"

type ChangeReqCreateRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Validate normalizes and verifies that all required fields are set.
// It returns true if the request is acceptable, false otherwise.
func (r *ChangeReqCreateRequest) Validate() bool {
	r.format()

	if len(r.Title) == 0 || len(r.Content) == 0 {
		return false
	}

	return true
}

// format cleans up the request fields by trimming leading and trailing
// whitespace from Name and Version. This helper is intended for internal use
// and is invoked automatically by the Validate method.
func (r *ChangeReqCreateRequest) format() {
	r.Title = strings.TrimSpace(r.Title)
}
