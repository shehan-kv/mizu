package changerequest

type EntryCreateRequest struct {
	Content string `json:"content"`
}

// Validate normalizes and verifies that all required fields are set.
// It returns true if the request is acceptable, false otherwise.
func (r *EntryCreateRequest) Validate() bool {

	return len(r.Content) != 0
}
