package store

// ContractStore abstracts all persistence operations for contracts.
// Lower-level errors should be wrapped in domain-specific error types
// defined in the store package.
// Implementations must be safe for concurrent use.
type ContractStore interface {
}
