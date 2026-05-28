package contract

import "errors"

var (
	ErrContractReplaceSignatoriesRequiresPending = errors.New("contract replace signatories requires pending")
	ErrContractMustHaveAtLeastTwoSignatories     = errors.New("contract must have at least two signatories")
	ErrContractSignatoryNotProjectMember         = errors.New("contract signatory not project member")
	ErrContractSignatoriesMustBeUnique           = errors.New("contract signatories must be unique")
	ErrContractMustHaveClientSignatory           = errors.New("contract must have client signatory")
	ErrContractInvalidSignatoryStatus            = errors.New("contract invalid signatory status")
	ErrContractConcurrentModification            = errors.New("contract concurrent modification")
	ErrContractSignatoryAlreadyActed             = errors.New("contract signatory already acted")
	ErrContractRejectRequiresPending             = errors.New("contract reject requires pending")
	ErrContractMustHaveTeamSignatory             = errors.New("contract must have team signatory")
	ErrContractSignRequiresPending               = errors.New("contract sign requires pending")
	ErrContractSignatoryNotFound                 = errors.New("contract signatory not found")
	ErrContractNameCannotBeEmpty                 = errors.New("contract name cannot be empty")
	ErrContractNameAlreadyExists                 = errors.New("contract name already exists")
	ErrContractSignatoriesLocked                 = errors.New("contract signatories locked")
	ErrContractIDCannotBeEmpty                   = errors.New("contract id cannot be empty")
	ErrContractInvalidStatus                     = errors.New("contract invalid status")
	ErrContractTermsTooShort                     = errors.New("contract terms too short")
	ErrContractNotFound                          = errors.New("contract not found")
)
