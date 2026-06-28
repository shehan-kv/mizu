package task

import (
	"context"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

type Service struct {
	userRepo    iam.UserRepository
	projectRepo project.Repository
}

func NewService(userRepo iam.UserRepository, projectRepo project.Repository) *Service {
	return &Service{userRepo: userRepo, projectRepo: projectRepo}
}

func (s *Service) EnsureValidAssignees(
	ctx context.Context,
	assignees []iam.UserID,
	project *project.Project) error {

	allExist, err := s.userRepo.ExistsAll(ctx, assignees)
	if err != nil {
		return err
	}

	if !allExist {
		return iam.ErrUserNotFound
	}

	for i := range assignees {
		if !project.HasMember(assignees[i]) {
			return ErrTaskAssigneeNotProjectMember
		}
	}

	return nil
}
