package project

import (
	"context"
	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/application/shared"
	"mizu/internal/application/uow"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"mizu/internal/domain/task"
	"time"
)

type Service struct {
	userRepo     iam.UserRepository
	projectRepo  project.Repository
	taskRepo     task.Repository
	billingRepo  billing.Repository
	contractRepo contract.Repository
	fileRepo     message.FileRepository

	uow uow.UnitOfWork

	authzSrv   *authz.Service
	projectSrv *project.Service

	internalBus eventbus.InternalBus

	idGen  common.IDGenerator
	logger logger.Logger
}

func NewService(
	userRepo iam.UserRepository,
	projectRepo project.Repository,
	taskRepo task.Repository,
	billingRepo billing.Repository,
	contractRepo contract.Repository,
	fileRepo message.FileRepository,
	uow uow.UnitOfWork,
	authzSrv *authz.Service,
	projectSrv *project.Service,
	internalBus eventbus.InternalBus,
	idGen common.IDGenerator,
	logger logger.Logger,
) *Service {
	return &Service{
		userRepo:     userRepo,
		projectRepo:  projectRepo,
		taskRepo:     taskRepo,
		billingRepo:  billingRepo,
		contractRepo: contractRepo,
		fileRepo:     fileRepo,
		uow:          uow,
		authzSrv:     authzSrv,
		projectSrv:   projectSrv,
		internalBus:  internalBus,
		idGen:        idGen,
		logger:       logger,
	}
}

func (s *Service) ReplaceMemberProjects(ctx context.Context, memberID string, actorID string, projectIDs []string) error {

	now := time.Now()

	mID, err := iam.NewUserID(memberID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	err = s.authzSrv.RequireAdministrator(ctx, actor)
	if err != nil {
		return err
	}

	exists, err := s.userRepo.Exists(ctx, mID)
	if err != nil {
		return err
	}

	if !exists {
		return iam.ErrUserNotFound
	}

	currentProjects, err := s.projectRepo.ListByMember(ctx, mID)
	if err != nil {
		return err
	}

	for _, p := range currentProjects {
		p.RemoveMember(mID, now)
	}

	ids := make([]project.ProjectID, len(projectIDs))
	for i := range projectIDs {
		pID, err := project.NewProjectID(projectIDs[i])
		if err != nil {
			return err
		}

		ids[i] = pID
	}

	newProjects, err := s.projectSrv.AssignMemberToProjects(ctx, mID, ids, now)
	if err != nil {
		return err
	}

	seen := make(map[project.ProjectID]struct{})
	affected := make([]*project.Project, 0)

	for _, p := range append(currentProjects, newProjects...) {
		if _, exists := seen[p.ID()]; !exists {
			seen[p.ID()] = struct{}{}
			affected = append(affected, p)
		}
	}

	return s.projectRepo.SaveAll(ctx, affected)
}

func (s *Service) CreateProject(ctx context.Context, params CreateProjectParams) error {

	now := time.Now()

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	err = s.authzSrv.RequireAdministrator(ctx, actor)
	if err != nil {
		return err
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	pID, err := project.NewProjectID(id)
	if err != nil {
		return err
	}

	name, err := project.NewName(params.Name)
	if err != nil {
		return err
	}

	projectStatus, err := project.NewStatus(params.Status)
	if err != nil {
		return err
	}

	memIDs := make([]iam.UserID, 0, len(params.Members))
	memIDs = append(memIDs, actor)
	for i := range params.Members {
		memIDs = append(memIDs, iam.UserID(params.Members[i]))

	}

	allExist, err := s.userRepo.ExistsAll(ctx, memIDs)
	if err != nil {
		return err
	}
	if !allExist {
		return iam.ErrUserNotFound
	}

	p, err := project.NewProject(
		pID,
		name,
		projectStatus,
		actor,
		memIDs,
		now,
	)
	if err != nil {
		return err
	}

	err = s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.projectRepo.Add(ctx, p)
	})
	if err != nil {
		return err
	}

	s.publishEvents(ctx, p.PullEvents())

	return nil
}

func (s *Service) ListStats(ctx context.Context, params ListStatsParams) (*shared.Collection[StatsDTO], error) {

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return nil, err
	}

	memID, err := iam.NewUserID(params.MemberID)
	if err != nil {
		return nil, err
	}

	if err := s.authzSrv.RequireAdministratorOrSelf(ctx, actor, memID); err != nil {
		return nil, err
	}

	filter := project.Filter{
		Keyword:  params.Keyword,
		MemberID: memID,
	}

	if params.Status != nil {
		projectStatus, err := project.NewStatus(*params.Status)
		if err != nil {
			return nil, err
		}
		filter.Status = &projectStatus
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	p, err := s.projectRepo.List(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	pIDs := make([]project.ProjectID, len(p))
	for i := range p {
		pIDs[i] = p[i].ID()
	}

	taskStats, err := s.taskRepo.ListStatsByProjects(ctx, pIDs)
	if err != nil {
		return nil, err
	}

	billingStats, err := s.billingRepo.ListStatsByProjects(ctx, pIDs)
	if err != nil {
		return nil, err
	}

	dto := make([]StatsDTO, len(p))
	for i := range p {
		ts, ok := taskStats[p[i].ID()]
		if !ok {
			s.logger.Warn("missing task stats", "project_id", p[i].ID())
			ts = task.Stats{}
		}

		bs, ok := billingStats[p[i].ID()]
		if !ok {
			s.logger.Warn("missing billing stats", "project_id", p[i].ID())
			bs = billing.Stats{}
		}

		dto[i] = StatsDTO{
			ID:             p[i].ID().String(),
			Name:           p[i].Name().String(),
			Status:         p[i].Status().String(),
			CreatedAt:      p[i].CreatedAt(),
			TotalTasks:     ts.Total(),
			TasksCompleted: ts.Completed(),
			TotalInvoices:  bs.InvoiceCount(),
			InvoicesPaid:   bs.InvoicePaidCount(),
			TotalQuotes:    bs.QuoteCount(),
		}
	}

	count, err := s.projectRepo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &shared.Collection[StatsDTO]{
		Items:      dto,
		TotalCount: count,
	}, nil

}

func (s *Service) ListAllStats(ctx context.Context, memberID string, actorID string) ([]StatsDTO, error) {

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	memID, err := iam.NewUserID(memberID)
	if err != nil {
		return nil, err
	}

	if err := s.authzSrv.RequireAdministratorOrSelf(ctx, actor, memID); err != nil {
		return nil, err
	}

	p, err := s.projectRepo.ListByMember(ctx, memID)
	if err != nil {
		return nil, err
	}

	pIDs := make([]project.ProjectID, len(p))
	for i := range p {
		pIDs[i] = p[i].ID()
	}

	taskStats, err := s.taskRepo.ListStatsByProjects(ctx, pIDs)
	if err != nil {
		return nil, err
	}

	billingStats, err := s.billingRepo.ListStatsByProjects(ctx, pIDs)
	if err != nil {
		return nil, err
	}

	dto := make([]StatsDTO, len(p))
	for i := range p {
		ts, ok := taskStats[p[i].ID()]
		if !ok {
			s.logger.Warn("missing task stats", "project_id", p[i].ID())
			ts = task.Stats{}
		}

		bs, ok := billingStats[p[i].ID()]
		if !ok {
			s.logger.Warn("missing billing stats", "project_id", p[i].ID())
			bs = billing.Stats{}
		}

		dto[i] = StatsDTO{
			ID:             p[i].ID().String(),
			Name:           p[i].Name().String(),
			Status:         p[i].Status().String(),
			CreatedAt:      p[i].CreatedAt(),
			TotalTasks:     ts.Total(),
			TasksCompleted: ts.Completed(),
			TotalInvoices:  bs.InvoiceCount(),
			InvoicesPaid:   bs.InvoicePaidCount(),
			TotalQuotes:    bs.QuoteCount(),
		}
	}

	return dto, nil

}

func (s *Service) GetProjectOverview(ctx context.Context, projectID string, actorID string) (ProjectOverviewDTO, error) {

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	if !p.HasMember(actor) {
		return ProjectOverviewDTO{}, project.ErrNotProjectMember
	}

	taskStats, err := s.taskRepo.GetStatsByProject(ctx, pID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	billingStats, err := s.billingRepo.GetStatsByProject(ctx, pID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	contractStats, err := s.contractRepo.GetStatsByProject(ctx, pID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	fileStats, err := s.fileRepo.GetStatsByProject(ctx, pID)
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	users, err := s.userRepo.ListByIDs(ctx, p.Members(), iam.UserFilter{})
	if err != nil {
		return ProjectOverviewDTO{}, err
	}

	members := make([]MemberDTO, len(users))
	for i, u := range users {
		members[i] = MemberDTO{
			ID:        u.ID().String(),
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			HasImage:  u.Image() != nil,
			Role:      u.Role().String(),
		}
	}

	return ProjectOverviewDTO{
		ID:                  p.ID().String(),
		Name:                p.Name().String(),
		Status:              p.Status().String(),
		CreatedAt:           p.CreatedAt(),
		Members:             members,
		TaskCount:           taskStats.Total(),
		TaskCompletedCount:  taskStats.Completed(),
		InvoiceCount:        billingStats.InvoiceCount(),
		InvoicePaidCount:    billingStats.InvoicePaidCount(),
		QuoteCount:          billingStats.QuoteCount(),
		ContractCount:       contractStats.Total(),
		ContractSignedCount: contractStats.Signed(),
		FileCount:           fileStats.FileCount(),
	}, nil
}

func (s *Service) GetCreatedCount(ctx context.Context, actorID string) ([]MetricDTO, error) {

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	metrics, err := s.projectRepo.ListCreatedPerDay(ctx, actor)
	if err != nil {
		return nil, err
	}

	dto := make([]MetricDTO, 0, len(metrics))

	for i := range metrics {
		dto = append(dto, MetricDTO{
			Key:   metrics[i].Key(),
			Value: metrics[i].Value(),
		})
	}

	return dto, nil
}

func (s *Service) ListMembers(ctx context.Context, params ListMembersParams) ([]MemberDTO, error) {

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return nil, err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return nil, err
	}

	if !p.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	members, err := s.userRepo.ListByIDs(ctx, p.Members(), iam.UserFilter{
		Keyword: params.Keyword,
	})
	if err != nil {
		return nil, err
	}

	dto := make([]MemberDTO, 0, len(members))

	for i := range members {
		dto = append(dto, MemberDTO{
			ID:        members[i].ID().String(),
			FirstName: members[i].FirstName(),
			LastName:  members[i].LastName(),
			Title:     members[i].Title(),
			HasImage:  members[i].Image() != nil,
			Role:      members[i].Role().String(),
		})
	}

	return dto, nil

}

func (s *Service) ReplaceMembers(ctx context.Context, projectID string, memberIDs []string, actorID string) error {

	now := time.Now()

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministrator(ctx, actor); err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	ids := make([]iam.UserID, len(memberIDs))
	for i := range memberIDs {
		id, err := iam.NewUserID(memberIDs[i])
		if err != nil {
			return err
		}

		ids[i] = id
	}

	allExist, err := s.userRepo.ExistsAll(ctx, ids)
	if err != nil {
		return err
	}
	if !allExist {
		return iam.ErrUserNotFound
	}

	p.ReplaceMembers(ids, now)

	if err := s.projectSrv.EnsureHasAdministrator(ctx, p); err != nil {
		return err
	}

	return s.projectRepo.Save(ctx, p)
}

func (s *Service) StartProject(ctx context.Context, projectID string, actorID string) error {

	now := time.Now()

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	p.MarkAsStarted(now)

	return s.projectRepo.Save(ctx, p)
}

func (s *Service) PauseProject(ctx context.Context, projectID string, actorID string) error {

	now := time.Now()

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	p.MarkAsPaused(now)

	return s.projectRepo.Save(ctx, p)
}

func (s *Service) CancelProject(ctx context.Context, projectID string, actorID string) error {

	now := time.Now()

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	p.MarkAsCancelled(now)

	return s.projectRepo.Save(ctx, p)
}

func (s *Service) CompleteProject(ctx context.Context, projectID string, actorID string) error {

	now := time.Now()

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	p.MarkAsCompleted(now)

	return s.projectRepo.Save(ctx, p)
}

func (s *Service) Delete(ctx context.Context, projectID string, actorID string) error {

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministrator(ctx, actor); err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	return s.projectRepo.Remove(ctx, p)
}

func (s *Service) publishEvents(ctx context.Context, events []common.Event) {
	for _, event := range events {
		if err := s.internalBus.Publish(ctx, event); err != nil {
			s.logger.Warn("failed to publish event",
				"event_type", event.EventType(),
				"err", err,
			)
		}
	}
}
