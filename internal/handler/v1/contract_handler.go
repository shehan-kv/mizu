package v1

import "mizu/internal/service"

// ContractHandler provides HTTP handlers for contract-related endpoints.
type ContractHandler struct {
	contSrv *service.ContractService
}

// NewContractHandler constructs a new ContractHandler.
func NewContractHandler(contSrv *service.ContractService) *ContractHandler {
	return &ContractHandler{
		contSrv: contSrv,
	}
}
