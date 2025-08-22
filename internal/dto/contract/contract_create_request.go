package contract

import "strings"

// ContractCreateRequest represents a contract create request
type ContractCreateRequest struct {
	// Name is the human-readable identifier for the contract.
	Name string `json:"name"`

	// Version designates the contract’s version string.
	Version string `json:"version"`

	// Contract contains the contract definition text.
	Contract string `json:"contract"`
}

// Validate normalizes and verifies that all required fields are set.
// It returns true if the request is acceptable, false otherwise.
func (r *ContractCreateRequest) Validate() bool {
	r.format()

	if len(r.Name) == 0 || len(r.Version) == 0 || len(r.Contract) == 0 {
		return false
	}

	return true
}

// format cleans up the request fields by trimming leading and trailing
// whitespace from Name and Version. This helper is intended for internal use
// and is invoked automatically by the Validate method.
func (r *ContractCreateRequest) format() {
	r.Name = strings.TrimSpace(r.Name)
	r.Version = strings.TrimSpace(r.Version)
}
