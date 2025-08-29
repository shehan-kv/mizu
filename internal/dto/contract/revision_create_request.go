package contract

import "strings"

type RevisionCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Validate normalizes and verifies that all required fields are set.
// It returns true if the request is acceptable, false otherwise.
func (r *RevisionCreateRequest) Validate() bool {
	r.format()

	if len(r.Title) == 0 || len(r.Description) == 0 {
		return false
	}

	return true
}

// format cleans up the request fields by trimming leading and trailing
// whitespace from Title and Description. This helper is intended for internal use
// and is invoked automatically by the Validate method.
func (r *RevisionCreateRequest) format() {
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
}
