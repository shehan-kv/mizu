package contract

import (
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ValidateSignatories(users []*iam.User, p *project.Project) error {
	hasClient := false
	hasTeamMember := false
	for i := range users {
		if !p.HasMember(users[i].ID()) {
			return ErrContractSignatoryNotProjectMember
		}

		if users[i].IsClient() {
			hasClient = true
		}

		if users[i].IsAdministrator() || users[i].IsStaff() {
			hasTeamMember = true
		}
	}
	if !hasClient {
		return ErrContractMustHaveClientSignatory
	}
	if !hasTeamMember {
		return ErrContractMustHaveTeamSignatory
	}
	return nil
}
