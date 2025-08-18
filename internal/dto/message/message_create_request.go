package message

// Represents a project message create request
type MessageCreateRequest struct {
	Message string `json:"message"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (r *MessageCreateRequest) Validate() bool {
	return len(r.Message) != 0
}
