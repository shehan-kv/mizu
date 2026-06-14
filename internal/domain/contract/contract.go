package contract

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"time"
)

type Contract struct {
	id          ContractID
	projectID   project.ProjectID
	name        Name
	status      Status
	terms       Terms
	signatories []Signatory

	events []common.Event

	version   int
	createdAt time.Time
	updatedAt time.Time
}

func NewContract(
	id ContractID,
	projectID project.ProjectID,
	name Name,
	terms Terms,
	signatories []Signatory,
	now time.Time,
) (*Contract, error) {

	if len(signatories) < 2 {
		return nil, ErrContractMustHaveAtLeastTwoSignatories
	}

	if hasDuplicateSignatories(signatories) {
		return nil, ErrContractSignatoriesMustBeUnique
	}

	c := Contract{
		id:          id,
		projectID:   projectID,
		name:        name,
		status:      StatusPending,
		terms:       terms,
		signatories: signatories,

		version:   1,
		createdAt: now,
		updatedAt: now,
	}

	c.events = append(c.events, ContractCreatedEvent{
		ContractID: id,
		ProjectID:  projectID,
		Name:       name,
		OccurredAt: now,
	})

	return &c, nil
}

func RestoreContract(
	id ContractID,
	projectID project.ProjectID,
	name Name,
	status Status,
	terms Terms,
	signatories []Signatory,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) *Contract {
	return &Contract{
		id:          id,
		projectID:   projectID,
		name:        name,
		status:      status,
		terms:       terms,
		signatories: signatories,

		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func hasDuplicateSignatories(signatories []Signatory) bool {
	seen := make(map[iam.UserID]struct{}, len(signatories))
	for _, s := range signatories {
		if _, exists := seen[s.userID]; exists {
			return true
		}
		seen[s.userID] = struct{}{}
	}
	return false
}

func (c *Contract) ID() ContractID {
	return c.id
}

func (c *Contract) ProjectID() project.ProjectID {
	return c.projectID
}

func (c *Contract) Name() Name {
	return c.name
}

func (c *Contract) Status() Status {
	return c.status
}

func (c *Contract) Terms() Terms {
	return c.terms
}

func (c *Contract) Signatories() []Signatory {
	signatories := make([]Signatory, len(c.signatories))
	copy(signatories, c.signatories)
	return signatories
}

func (c *Contract) Sign(userID iam.UserID, now time.Time) error {

	if c.status != StatusPending {
		return ErrContractSignRequiresPending
	}

	idx, err := c.findPendingSignatory(userID)
	if err != nil {
		return err
	}

	c.signatories[idx].status = SignatoryStatusSigned
	c.signatories[idx].updatedAt = now

	allSigned := true
	for i := range c.signatories {
		if c.signatories[i].status != SignatoryStatusSigned {
			allSigned = false
			break
		}
	}
	if allSigned {
		c.status = StatusSigned
	}

	c.updatedAt = now

	c.events = append(c.events, ContractStatusChangedEvent{
		ContractID: c.id,
		ProjectID:  c.projectID,
		Name:       c.name,
		Signatory:  c.signatories[idx],
		Status:     c.status,
		OccurredAt: now,
	})

	return nil
}

func (c *Contract) HasSignatory(userID iam.UserID) bool {
	for _, s := range c.Signatories() {
		if s.UserID() == userID {
			return true
		}
	}
	return false
}

func (c *Contract) Reject(userID iam.UserID, now time.Time) error {
	if c.status != StatusPending {
		return ErrContractRejectRequiresPending
	}

	idx, err := c.findPendingSignatory(userID)
	if err != nil {
		return err
	}

	c.signatories[idx].status = SignatoryStatusRejected
	c.signatories[idx].updatedAt = now
	c.status = StatusRejected
	c.updatedAt = now

	c.events = append(c.events, ContractStatusChangedEvent{
		ContractID: c.id,
		ProjectID:  c.projectID,
		Name:       c.name,
		Signatory:  c.signatories[idx],
		Status:     c.status,
		OccurredAt: now,
	})

	return nil
}

func (c *Contract) findPendingSignatory(userID iam.UserID) (int, error) {
	for i := range c.signatories {
		if c.signatories[i].userID == userID {
			if c.signatories[i].status != SignatoryStatusPending {
				return -1, ErrContractSignatoryAlreadyActed
			}
			return i, nil
		}
	}
	return -1, ErrContractSignatoryNotFound
}

func (c *Contract) ReplaceSignatories(userIDs []iam.UserID, now time.Time) error {

	if c.status != StatusPending {
		return ErrContractReplaceSignatoriesRequiresPending
	}

	if len(userIDs) < 2 {
		return ErrContractMustHaveAtLeastTwoSignatories
	}

	if hasDuplicateUserIDs(userIDs) {
		return ErrContractSignatoriesMustBeUnique
	}

	for i := range c.signatories {
		if c.signatories[i].status != SignatoryStatusPending {
			return ErrContractSignatoriesLocked
		}
	}

	signatories := make([]Signatory, len(userIDs))
	for i := range userIDs {
		signatories[i] = Signatory{userID: userIDs[i], status: SignatoryStatusPending}
	}
	c.signatories = signatories

	c.updatedAt = now
	return nil
}

func hasDuplicateUserIDs(userIDs []iam.UserID) bool {
	seen := make(map[iam.UserID]struct{}, len(userIDs))
	for i := range userIDs {
		if _, exists := seen[userIDs[i]]; exists {
			return true
		}
		seen[userIDs[i]] = struct{}{}
	}
	return false
}

func (c *Contract) Version() int {
	return c.version
}

func (c *Contract) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Contract) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Contract) PullEvents() []common.Event {
	events := c.events
	c.events = nil

	return events
}

func (c *Contract) Equals(other *Contract) bool {
	return c.id == other.id
}
