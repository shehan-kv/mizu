package project

import (
	"context"
	"mizu/internal/domain/iam"
	"time"
)

type Service struct {
	userRepo    iam.UserRepository
	projectRepo Repository
}

func NewService(userRepo iam.UserRepository, projectRepo Repository) *Service {
	return &Service{userRepo: userRepo, projectRepo: projectRepo}
}

func (s *Service) EnsureHasAdministrator(ctx context.Context, project *Project) error {

	ok, err := s.userRepo.IsAnyAdministrator(ctx, project.Members())
	if err != nil {
		return err
	}

	if !ok {
		return ErrProjectMustHaveAdministrator
	}

	return nil
}

func (s *Service) AssignMemberToProjects(
	ctx context.Context,
	memberID iam.UserID,
	projectIDs []ProjectID, now time.Time) ([]*Project, error) {

	projects, err := s.projectRepo.ListByIDs(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	if len(projects) != len(projectIDs) {
		return nil, ErrProjectNotFound
	}

	for i := range projects {
		projects[i].AddMember(memberID, now)
	}

	return projects, nil
}
