package task

import (
	"context"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

type Service struct {
	iamRepo     iam.Repository
	projectRepo project.Repository
}

func NewService(iamRepo iam.Repository, projectRepo project.Repository) *Service {
	return &Service{iamRepo: iamRepo, projectRepo: projectRepo}
}

func (s *Service) EnsureValidAssignees(
	ctx context.Context,
	assignees []iam.UserID,
	project *project.Project) error {

	allExist, err := s.iamRepo.ExistsAll(ctx, assignees)
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
