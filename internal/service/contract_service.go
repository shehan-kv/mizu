package service

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
)

// ContractService handles contract-related business logic.
// It relies on the provided ContractStore for database operations
// and uses the Logger for audit and debugging.
type ContractService struct {
	lg     logger.Logger
	contSt store.ContractStore
}

// NewContractService constructs a ContractService that
// handles contract-related business logic. It needs a non-nil logger
// for audit and debugging, and a ContractStore for persistence.
func NewContractService(lg logger.Logger, contSt store.ContractStore) *ContractService {
	return &ContractService{
		lg:     lg,
		contSt: contSt,
	}
}
