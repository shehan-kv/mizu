package authz

import (
	"context"
	"mizu/internal/domain/iam"
)

type Service struct {
	iamRepo iam.Repository
}

func NewService(iamRepo iam.Repository) *Service {
	return &Service{iamRepo: iamRepo}
}

func (s *Service) RequireAdministrator(ctx context.Context, actorID iam.UserID) error {
	actor, err := s.iamRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}
	if !actor.IsAdministrator() {
		return ErrForbidden
	}
	return nil
}

func (s *Service) RequireAdministratorOrSelf(ctx context.Context, actorID iam.UserID, targetID iam.UserID) error {
	if actorID == targetID {
		return nil
	}
	return s.RequireAdministrator(ctx, actorID)
}

func (s *Service) RequireAdministratorOrStaff(ctx context.Context, actorID iam.UserID) error {
	actor, err := s.iamRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}
	if !actor.IsAdministrator() && !actor.IsStaff() {
		return ErrForbidden
	}
	return nil
}
func (s *Service) RequireAdministratorStaffOrSelf(ctx context.Context, actorID iam.UserID, targetID iam.UserID) error {
	if actorID == targetID {
		return nil
	}
	return s.RequireAdministratorOrStaff(ctx, actorID)
}

func (s *Service) RequireClient(ctx context.Context, actorID iam.UserID) error {
	actor, err := s.iamRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}
	if !actor.IsClient() {
		return ErrForbidden
	}
	return nil
}
