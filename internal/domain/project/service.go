package project

import (
	"context"
	"mizu/internal/domain/iam"
	"time"
)

type Service struct {
	iamRepo     iam.Repository
	projectRepo Repository
}

func NewService(iamRepo iam.Repository, projectRepo Repository) *Service {
	return &Service{iamRepo: iamRepo, projectRepo: projectRepo}
}

func (s *Service) EnsureHasAdministrator(ctx context.Context, project *Project) error {

	ok, err := s.iamRepo.IsAnyAdministrator(ctx, project.Members())
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
