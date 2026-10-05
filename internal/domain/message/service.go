package message

import "mizu/internal/domain/iam"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ValidateStandaloneChannelMembers(users []*iam.User) error {
	for _, u := range users {
		if u.IsAdministrator() || u.IsStaff() {
			return nil
		}
	}
	return ErrChannelMustHaveStaffMember
}
