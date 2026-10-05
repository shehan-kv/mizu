package contract

import (
	"context"
	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/application/shared"
	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"time"
)

type Service struct {
	userRepo     iam.UserRepository
	contractRepo contract.Repository
	projectRepo  project.Repository

	contractSrv *contract.Service
	authzSrv    *authz.Service

	internalBus eventbus.InternalBus

	idGen  common.IDGenerator
	mailer mailer.Mailer

	logger logger.Logger
}

func NewService(
	userRepo iam.UserRepository,
	contractRepo contract.Repository,
	projectRepo project.Repository,
	contractSrv *contract.Service,
	authzSrv *authz.Service,
	internalBus eventbus.InternalBus,
	idGen common.IDGenerator,
	mailer mailer.Mailer,
	logger logger.Logger,
) *Service {
	return &Service{
		userRepo:     userRepo,
		contractRepo: contractRepo,
		projectRepo:  projectRepo,
		contractSrv:  contractSrv,
		authzSrv:     authzSrv,
		internalBus:  internalBus,
		idGen:        idGen,
		mailer:       mailer,
		logger:       logger,
	}
}

func (s *Service) CreateContract(ctx context.Context, params CreateContractParams) error {

	now := time.Now()

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	name, err := contract.NewName(params.Name)
	if err != nil {
		return err
	}

	terms, err := contract.NewTerms(params.Terms)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	signIDs := make([]iam.UserID, len(params.SignatoryIDs))
	for i := range params.SignatoryIDs {
		id, err := iam.NewUserID(params.SignatoryIDs[i])
		if err != nil {
			return err
		}
		signIDs[i] = id
	}

	users, err := s.userRepo.ListByIDs(ctx, signIDs, iam.UserFilter{})
	if err != nil {
		return err
	}

	if err := s.contractSrv.ValidateSignatories(users, p); err != nil {
		return err
	}

	signs := make([]contract.Signatory, len(signIDs))
	for i := range params.SignatoryIDs {
		signs[i] = contract.NewSignatory(signIDs[i], now)
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	cID, err := contract.NewContractID(id)
	if err != nil {
		return err
	}

	c, err := contract.NewContract(cID, pID, name, terms, signs, now)
	if err != nil {
		return err
	}

	if err := s.contractRepo.Add(ctx, c); err != nil {
		return err
	}

	s.publishEvents(ctx, c.PullEvents())

	return nil
}

func (s *Service) Sign(ctx context.Context, actorID string, contractID string) error {

	now := time.Now()

	cID, err := contract.NewContractID(contractID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	c, err := s.contractRepo.Get(ctx, cID)
	if err != nil {
		return err
	}

	if err := c.Sign(actor, now); err != nil {
		return err
	}

	if err := s.contractRepo.Save(ctx, c); err != nil {
		return err
	}

	s.publishEvents(ctx, c.PullEvents())

	return nil
}

func (s *Service) Reject(ctx context.Context, actorID string, contractID string) error {

	now := time.Now()

	cID, err := contract.NewContractID(contractID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	c, err := s.contractRepo.Get(ctx, cID)
	if err != nil {
		return err
	}

	if err := c.Reject(actor, now); err != nil {
		return err
	}

	if err := s.contractRepo.Save(ctx, c); err != nil {
		return err
	}

	s.publishEvents(ctx, c.PullEvents())

	return nil
}

func (s *Service) GetContract(ctx context.Context, actorID string, contractID string) (*ContractDTO, error) {

	cID, err := contract.NewContractID(contractID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	c, err := s.contractRepo.Get(ctx, cID)
	if err != nil {
		return nil, err
	}

	p, err := s.projectRepo.Get(ctx, c.ProjectID())
	if err != nil {
		return nil, err
	}

	if !p.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	signs := c.Signatories()

	signIDs := make([]iam.UserID, 0, len(signs))
	for i := range signs {
		signIDs = append(signIDs, signs[i].UserID())
	}

	users, err := s.userRepo.ListByIDs(ctx, signIDs, iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	userMap := make(map[iam.UserID]*iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = users[i]
	}

	signDTO := make([]SignatoryDTO, 0, len(signs))

	var memberSignatoryStatus string

	for i := range signs {
		u, ok := userMap[signs[i].UserID()]
		if !ok {
			return nil, contract.ErrContractSignatoryNotFound
		}

		if signs[i].UserID() == actor {
			memberSignatoryStatus = signs[i].Status().String()
		}

		signDTO = append(signDTO, SignatoryDTO{
			ID:        u.ID().String(),
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			Title:     u.Title(),
			Role:      u.Role().String(),
			HasImage:  u.Image() != nil,
			Status:    signs[i].Status().String(),
			UpdatedAt: signs[i].UpdatedAt(),
		})
	}

	return &ContractDTO{
		ID:                    c.ID().String(),
		ProjectID:             c.ProjectID().String(),
		Name:                  c.Name().String(),
		Status:                c.Status().String(),
		Terms:                 c.Terms().String(),
		MemberSignatoryStatus: memberSignatoryStatus,
		Signatories:           signDTO,
		CreatedAt:             c.CreatedAt(),
		UpdatedAt:             c.UpdatedAt(),
	}, nil
}

func (s *Service) ListOverviewByProject(ctx context.Context, params ListByProjectParams) (*shared.Collection[ContractOverviewDTO], error) {

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return nil, err
	}

	filter := contract.FilterByProject{
		ProjectID: pID,
		Keyword:   params.Keyword,
	}

	if params.Status != nil {
		status, err := contract.NewStatus(*params.Status)
		if err != nil {
			return nil, err
		}

		filter.Status = &status
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return nil, err
	}

	if !p.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	contracts, err := s.contractRepo.ListByProject(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	count, err := s.contractRepo.CountByProject(ctx, filter)
	if err != nil {
		return nil, err
	}

	userIDSet := make(map[iam.UserID]struct{})

	for i := range contracts {
		sg := contracts[i].Signatories()
		for j := range sg {
			userIDSet[sg[j].UserID()] = struct{}{}
		}
	}

	userIDs := make([]iam.UserID, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	users, err := s.userRepo.ListByIDs(ctx, userIDs, iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	userMap := make(map[iam.UserID]*iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = users[i]
	}

	items := make([]ContractOverviewDTO, 0, len(contracts))

	for i := range contracts {

		sg := contracts[i].Signatories()
		signatories := make([]SignatoryDTO, 0, len(sg))

		var memberSignatoryStatus string

		for j := range sg {

			u, ok := userMap[sg[j].UserID()]
			if !ok {
				return nil, contract.ErrContractSignatoryNotFound
			}

			if sg[j].UserID() == actor {
				memberSignatoryStatus = sg[j].Status().String()
			}

			signatories = append(signatories, SignatoryDTO{
				ID:        u.ID().String(),
				FirstName: u.FirstName(),
				LastName:  u.LastName(),
				Title:     u.Title(),
				Role:      u.Role().String(),
				HasImage:  u.Image() != nil,
				Status:    sg[j].Status().String(),
				UpdatedAt: sg[j].UpdatedAt(),
			})
		}

		items = append(items, ContractOverviewDTO{
			ID:                    contracts[i].ID().String(),
			ProjectID:             contracts[i].ProjectID().String(),
			Name:                  contracts[i].Name().String(),
			Status:                contracts[i].Status().String(),
			MemberSignatoryStatus: memberSignatoryStatus,
			Signatories:           signatories,
			CreatedAt:             contracts[i].CreatedAt(),
			UpdatedAt:             contracts[i].UpdatedAt(),
		})
	}

	return &shared.Collection[ContractOverviewDTO]{
		Items:      items,
		TotalCount: count,
	}, nil
}

func (s *Service) ListOverviewByMember(ctx context.Context, params ListByMemberParams) (*shared.Collection[ContractOverviewDTO], error) {

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

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	filter := contract.FilterBySignatory{
		SignatoryID: memID,
		Keyword:     params.Keyword,
	}

	if params.Status != nil {
		status, err := contract.NewStatus(*params.Status)
		if err != nil {
			return nil, err
		}

		filter.Status = &status
	}

	contracts, err := s.contractRepo.ListBySignatory(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	count, err := s.contractRepo.CountBySignatory(ctx, filter)
	if err != nil {
		return nil, err
	}

	userIDSet := make(map[iam.UserID]struct{})

	for i := range contracts {
		sg := contracts[i].Signatories()

		for j := range sg {
			userIDSet[sg[j].UserID()] = struct{}{}
		}
	}

	userIDs := make([]iam.UserID, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	users, err := s.userRepo.ListByIDs(ctx, userIDs, iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	userMap := make(map[iam.UserID]*iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = users[i]
	}

	items := make([]ContractOverviewDTO, 0, len(contracts))

	for i := range contracts {

		sg := contracts[i].Signatories()
		signatories := make([]SignatoryDTO, 0, len(sg))

		var memberSignatoryStatus string

		for j := range sg {

			u, ok := userMap[sg[j].UserID()]
			if !ok {
				return nil, contract.ErrContractSignatoryNotFound
			}

			if sg[j].UserID() == memID {
				memberSignatoryStatus = sg[j].Status().String()
			}

			signatories = append(signatories, SignatoryDTO{
				ID:        u.ID().String(),
				FirstName: u.FirstName(),
				LastName:  u.LastName(),
				Title:     u.Title(),
				Role:      u.Role().String(),
				HasImage:  u.Image() != nil,
				Status:    sg[j].Status().String(),
				UpdatedAt: sg[j].UpdatedAt(),
			})
		}

		items = append(items, ContractOverviewDTO{
			ID:                    contracts[i].ID().String(),
			ProjectID:             contracts[i].ProjectID().String(),
			Name:                  contracts[i].Name().String(),
			Status:                contracts[i].Status().String(),
			MemberSignatoryStatus: memberSignatoryStatus,
			Signatories:           signatories,
			CreatedAt:             contracts[i].CreatedAt(),
			UpdatedAt:             contracts[i].UpdatedAt(),
		})
	}

	return &shared.Collection[ContractOverviewDTO]{
		Items:      items,
		TotalCount: count,
	}, nil

}

func (s *Service) ListSignatories(ctx context.Context, contractID string, actorID string) ([]SignatoryDTO, error) {

	cID, err := contract.NewContractID(contractID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	c, err := s.contractRepo.Get(ctx, cID)
	if err != nil {
		return nil, err
	}

	if !c.HasSignatory(actor) {
		return nil, contract.ErrContractUserNotSignatory
	}

	signs := c.Signatories()

	signIDs := make([]iam.UserID, 0, len(signs))

	for _, s := range signs {
		uid := s.UserID()
		signIDs = append(signIDs, uid)
	}

	users, err := s.userRepo.ListByIDs(ctx, signIDs, iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*iam.User, len(users))

	for _, u := range users {
		userMap[u.ID().String()] = u
	}

	dto := make([]SignatoryDTO, 0, len(signs))
	for _, sign := range signs {
		uid := sign.UserID().String()

		u, ok := userMap[uid]
		if !ok {
			continue
		}

		dto = append(dto, SignatoryDTO{
			ID:        uid,
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			Title:     u.Title(),
			Role:      u.Role().String(),
			HasImage:  u.Image() != nil,
			Status:    sign.Status().String(),
			UpdatedAt: sign.UpdatedAt(),
		})
	}
	return dto, nil
}

func (s *Service) ReplaceSignatories(ctx context.Context, params ReplaceSignatoriesParams) error {
	now := time.Now()

	cID, err := contract.NewContractID(params.ContractID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	c, err := s.contractRepo.Get(ctx, cID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, c.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	signIDs := make([]iam.UserID, len(params.Signatories))
	for i, id := range params.Signatories {
		uID, err := iam.NewUserID(id)
		if err != nil {
			return err
		}
		signIDs[i] = uID
	}

	users, err := s.userRepo.ListByIDs(ctx, signIDs, iam.UserFilter{})
	if err != nil {
		return err
	}

	if err := s.contractSrv.ValidateSignatories(users, p); err != nil {
		return err
	}

	if err := c.ReplaceSignatories(signIDs, now); err != nil {
		return err
	}

	return s.contractRepo.Save(ctx, c)
}

func (s *Service) EmailContract(ctx context.Context, contractID string, actorID string, memberID string) error {

	cID, err := contract.NewContractID(contractID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	member, err := iam.NewUserID(memberID)
	if err != nil {
		return err
	}

	err = s.authzSrv.RequireAdministratorStaffOrSelf(ctx, actor, member)
	if err != nil {
		return err
	}

	c, err := s.contractRepo.Get(ctx, cID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, c.ProjectID())
	if err != nil {
		return err
	}

	m, err := s.userRepo.GetByID(ctx, member)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) || !p.HasMember(member) {
		return project.ErrNotProjectMember
	}

	signs := c.Signatories()

	signIDs := make([]iam.UserID, 0, len(signs))
	for i := range signs {
		signIDs = append(signIDs, signs[i].UserID())
	}

	users, err := s.userRepo.ListByIDs(ctx, signIDs, iam.UserFilter{})
	if err != nil {
		return err
	}

	userMap := make(map[iam.UserID]*iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = users[i]
	}

	signatories := make([]mailer.ContractSignatory, 0, len(signs))

	for i := range signs {
		u, ok := userMap[signs[i].UserID()]
		if !ok {
			return contract.ErrContractSignatoryNotFound
		}

		signatories = append(signatories, mailer.ContractSignatory{
			ID:        u.ID().String(),
			FirstName: u.FirstName(),
			LastName:  u.LastName(),
			Email:     u.Email().String(),
			Title:     u.Title(),
			Role:      u.Role().String(),
			Status:    signs[i].Status().String(),
			UpdatedAt: signs[i].UpdatedAt(),
		})
	}

	return s.mailer.SendContractEmail(ctx, mailer.ContractEmail{
		Subject:        "Contract",
		RecipientEmail: m.Email().String(),

		ContractID:    c.ID().String(),
		ContractName:  c.Name().String(),
		ContractTerms: c.Terms().String(),

		ProjectID:   p.ID().String(),
		ProjectName: p.Name().String(),

		Signatories: signatories,
	})
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
