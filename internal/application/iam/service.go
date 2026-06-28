package iam

import (
	"context"
	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/filestore"
	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/application/session"
	"mizu/internal/application/shared"
	"mizu/internal/application/uow"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"mizu/internal/domain/verification"
	"time"
)

type Service struct {
	userRepo         iam.UserRepository
	credRepo         iam.CredentialRepository
	verificationRepo verification.Repository
	projectRepo      project.Repository

	uow uow.UnitOfWork

	iamSrv     *iam.Service
	authzSrv   *authz.Service
	projectSrv *project.Service

	internalBus eventbus.InternalBus

	fileStore    filestore.Store
	idGen        common.IDGenerator
	sessionStore session.Store
	logger       logger.Logger
	mailer       mailer.Mailer
	pwHasher     iam.PasswordHasher
}

func NewService(
	userRepo iam.UserRepository,
	credRepo iam.CredentialRepository,
	verificationRepo verification.Repository,
	projectRepo project.Repository,
	uow uow.UnitOfWork,
	iamSrv *iam.Service,
	authzSrv *authz.Service,
	projectSrv *project.Service,
	internalBus eventbus.InternalBus,
	fileStore filestore.Store,
	idGen common.IDGenerator,
	sessionStore session.Store,
	logger logger.Logger,
	mailer mailer.Mailer,
	pwHasher iam.PasswordHasher,
) *Service {

	return &Service{
		userRepo:         userRepo,
		credRepo:         credRepo,
		verificationRepo: verificationRepo,
		projectRepo:      projectRepo,
		uow:              uow,
		iamSrv:           iamSrv,
		authzSrv:         authzSrv,
		projectSrv:       projectSrv,
		internalBus:      internalBus,
		fileStore:        fileStore,
		idGen:            idGen,
		sessionStore:     sessionStore,
		logger:           logger,
		mailer:           mailer,
		pwHasher:         pwHasher,
	}
}

func (s Service) EnsureDefaultAdminExists(ctx context.Context) error {

	now := time.Now()

	exists, err := s.userRepo.HasAdministrator(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	userID, err := iam.NewUserID(id)
	if err != nil {
		return err
	}

	name, err := iam.NewName("Default", "Administrator")
	if err != nil {
		return err
	}

	email, err := iam.NewEmail("admin@mizu")
	if err != nil {
		return err
	}

	pw, err := iam.NewPlainPassword("admin@mizu")
	if err != nil {
		return err
	}

	pwHash, err := s.pwHasher.Hash(pw)
	if err != nil {
		return err
	}

	cred := iam.NewCredential(userID, pwHash)

	admin := iam.NewSystemUser(userID, name, email, nil, iam.RoleAdministrator, true, now)

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Add(ctx, admin); err != nil {
			return err
		}
		return s.credRepo.Add(ctx, cred)
	})
}

func (s Service) SignIn(
	ctx context.Context,
	email string,
	password string,
	extSessionID *string) (SessionDTO, error) {

	parsedEmail, err := iam.NewEmail(email)
	if err != nil {
		return SessionDTO{}, err
	}

	plainPw, err := iam.NewPlainPassword(password)
	if err != nil {
		return SessionDTO{}, err
	}

	user, err := s.userRepo.GetByEmail(ctx, parsedEmail)
	if err != nil {
		return SessionDTO{}, err
	}

	err = s.iamSrv.Authenticate(ctx, user, plainPw)
	if err != nil {
		return SessionDTO{}, err
	}

	sID, err := s.idGen.Generate()
	if err != nil {
		return SessionDTO{}, err
	}

	sessionID, err := session.NewSessionID(sID)
	if err != nil {
		return SessionDTO{}, err
	}

	newSession := session.NewSession(sessionID, user.ID())

	if extSessionID != nil {
		if err := s.sessionStore.Delete(ctx, session.SessionID(*extSessionID)); err != nil {
			s.logger.Error("failed to revoke existing session",
				"sessionID", extSessionID,
				"err", err,
			)
		}

	}

	err = s.sessionStore.Add(ctx, newSession)
	if err != nil {
		return SessionDTO{}, err
	}

	user.RecordSignIn(time.Now())

	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("failed to record last sign in",
			"userID", user.ID(),
			"err", err,
		)
	}

	return SessionDTO{
		ID:        newSession.ID().String(),
		UserID:    newSession.UserID().String(),
		Role:      user.Role().String(),
		IssuedAt:  newSession.IssuedAt(),
		ExpiresAt: newSession.ExpiresAt(),
	}, nil
}

func (s Service) SignOut(ctx context.Context, sessionID string) error {
	sID, err := session.NewSessionID(sessionID)
	if err != nil {
		return nil
	}

	if err := s.sessionStore.Delete(ctx, sID); err != nil {
		return err
	}

	return nil
}

func (s Service) CreateUser(ctx context.Context, params CreateUserParams) error {

	now := time.Now()

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	err = s.authzSrv.RequireAdministrator(ctx, actor)
	if err != nil {
		return err
	}

	name, err := iam.NewName(params.FirstName, params.LastName)
	if err != nil {
		return err
	}

	parsedEmail, err := iam.NewEmail(params.Email)
	if err != nil {
		return err
	}

	role, err := iam.NewRole(params.Role)
	if err != nil {
		return err
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	userID, err := iam.NewUserID(id)
	if err != nil {
		return err
	}

	user := iam.NewUser(
		userID,
		name,
		parsedEmail,
		params.Title,
		role,
		params.IsActive,
		actor,
		now,
	)

	vID, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	verificationID, err := verification.NewVerificationID(vID)
	if err != nil {
		return err
	}

	verificationReq := verification.NewVerification(
		verificationID,
		user.ID(),
		now,
	)

	projectIDs := make([]project.ProjectID, len(params.ProjectIDs))
	for i := range params.ProjectIDs {
		pID, err := project.NewProjectID(params.ProjectIDs[i])
		if err != nil {
			return err
		}

		projectIDs[i] = pID
	}

	memberProjects, err := s.projectSrv.AssignMemberToProjects(ctx, user.ID(), projectIDs, now)
	if err != nil {
		return err
	}

	err = s.uow.Execute(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Add(ctx, user); err != nil {
			return err
		}
		if err := s.verificationRepo.Add(ctx, verificationReq); err != nil {
			return err
		}
		return s.projectRepo.SaveAll(ctx, memberProjects)
	})
	if err != nil {
		return err
	}

	s.publishEvents(ctx, user.PullEvents())
	s.publishEvents(ctx, verificationReq.PullEvents())

	return nil
}

func (s *Service) RegenerateVerification(ctx context.Context, userID string, actorID string) error {
	now := time.Now()

	uID, err := iam.NewUserID(userID)
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

	user, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return err
	}

	if user.IsVerified() {
		return iam.ErrUserAlreadyVerified
	}

	vID, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	verificationID, err := verification.NewVerificationID(vID)
	if err != nil {
		return err
	}

	req := verification.NewVerification(verificationID, user.ID(), now)

	err = s.uow.Execute(ctx, func(ctx context.Context) error {
		if err := s.verificationRepo.RemoveByUserID(ctx, user.ID()); err != nil {
			return err
		}
		return s.verificationRepo.Add(ctx, req)

	})
	if err != nil {
		return err
	}

	s.publishEvents(ctx, req.PullEvents())

	return nil
}

func (s *Service) VerifyAccount(ctx context.Context, verificationID string, password string) error {

	now := time.Now()

	vID, err := verification.NewVerificationID(verificationID)
	if err != nil {
		return err
	}

	verification, err := s.verificationRepo.Get(ctx, vID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, verification.UserID())
	if err != nil {
		return err
	}

	if user.IsVerified() {
		return iam.ErrUserAlreadyVerified
	}

	plainPw, err := iam.NewPlainPassword(password)
	if err != nil {
		return err
	}

	hash, err := s.pwHasher.Hash(plainPw)
	if err != nil {
		return err
	}

	cred := iam.NewCredential(user.ID(), hash)

	user.Verify(now)

	err = s.uow.Execute(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Save(ctx, user); err != nil {
			return err
		}
		if err := s.credRepo.Add(ctx, cred); err != nil {
			return err
		}
		return s.verificationRepo.Remove(ctx, verification)
	})
	if err != nil {
		return err
	}

	s.publishEvents(ctx, user.PullEvents())

	return nil
}

func (s *Service) VerificationExists(ctx context.Context, verificationID string) error {

	vID, err := verification.NewVerificationID(verificationID)
	if err != nil {
		return err
	}

	_, err = s.verificationRepo.Get(ctx, vID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Activate(ctx context.Context, userID string, actorID string) error {

	now := time.Now()

	uID, err := iam.NewUserID(userID)
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

	user, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return err
	}

	if user.IsActive() {
		return nil
	}

	user.Activate(now)

	return s.userRepo.Save(ctx, user)
}

func (s *Service) Deactivate(ctx context.Context, userID string, actorID string) error {

	now := time.Now()

	uID, err := iam.NewUserID(userID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if uID == actor {
		return iam.ErrUserCannotDeleteSelf
	}

	if err := s.authzSrv.RequireAdministrator(ctx, actor); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return err
	}

	if !user.IsActive() {
		return nil
	}

	user.Deactivate(now)

	return s.userRepo.Save(ctx, user)
}

func (s *Service) Update(ctx context.Context, params UpdateUserParams) error {
	now := time.Now()

	actorID, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	userID, err := iam.NewUserID(params.UserID)
	if err != nil {
		return err
	}

	err = s.authzSrv.RequireAdministratorOrSelf(ctx, actorID, userID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	name, err := iam.NewName(params.FirstName, params.LastName)
	if err != nil {
		return err
	}

	email, err := iam.NewEmail(params.Email)
	if err != nil {
		return err
	}

	user.ChangeName(name, now)
	user.ChangeEmail(email, now)
	user.ChangeTitle(params.Title, now)

	if actorID != userID {
		role, err := iam.NewRole(params.Role)
		if err != nil {
			return err
		}

		user.AssignRole(role, now)
	}

	if params.Image != nil {

		imageIDString, err := s.idGen.Generate()
		if err != nil {
			return err
		}

		imageName, err := iam.NewImageName(imageIDString)
		if err != nil {
			return err
		}

		mimeType, err := iam.NewMimeType(params.Image.MimeType)
		if err != nil {
			return err
		}

		img := iam.NewImage(imageName, mimeType)
		prevImg := user.Image()
		user.ChangeImage(&img, now)

		if prevImg != nil {
			if err := s.fileStore.Delete(ctx, prevImg.Name().String()); err != nil {
				return err
			}
		}

		if err := s.fileStore.Save(ctx, imageIDString, params.Image.Reader); err != nil {
			return err
		}

	}

	err = s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.userRepo.Save(ctx, user)
	})
	if err != nil {
		return err
	}

	s.publishEvents(ctx, user.PullEvents())

	return nil
}

func (s *Service) Delete(ctx context.Context, userID string, actorID string) error {

	uID, err := iam.NewUserID(userID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if uID == actor {
		return iam.ErrUserCannotDeleteSelf
	}

	if err := s.authzSrv.RequireAdministrator(ctx, actor); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return err
	}

	return s.userRepo.Remove(ctx, user)
}

func (s *Service) GetUser(ctx context.Context, actorID string, userID string) (*UserDTO, error) {

	uID, err := iam.NewUserID(userID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	if err := s.authzSrv.RequireAdministratorOrSelf(ctx, actor, uID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return nil, err
	}

	dto := s.toUserDTO(user)

	return &dto, nil
}

func (s *Service) GetProfileImage(ctx context.Context, userID string) (*ProfileImageDTO, error) {

	uID, err := iam.NewUserID(userID)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return nil, err
	}

	img := u.Image()

	if img == nil {
		return nil, iam.ErrUserImageNotFound
	}

	reader, err := s.fileStore.Open(ctx, img.Name().String())
	if err != nil {
		return nil, err
	}

	return &ProfileImageDTO{
		Name:     img.Name().String(),
		MimeType: img.MimeType().String(),
		Reader:   reader,
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, params ListUsersParams) (*shared.Collection[UserDTO], error) {

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return nil, err
	}

	if err := s.authzSrv.RequireAdministrator(ctx, actor); err != nil {
		return nil, err
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	filter := iam.UserFilter{
		Keyword:    params.Keyword,
		IsActive:   params.IsActive,
		IsVerified: params.IsVerified,
	}

	if params.Role != nil {
		role, err := iam.NewRole(*params.Role)
		if err != nil {
			return nil, err
		}
		filter.Role = &role
	}

	users, err := s.userRepo.List(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	count, err := s.userRepo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	dtos := make([]UserDTO, len(users))

	for i := range users {
		dtos[i] = s.toUserDTO(users[i])
	}

	resp := shared.Collection[UserDTO]{
		Items:      dtos,
		TotalCount: count,
	}

	return &resp, nil
}

func (s Service) toUserDTO(user *iam.User) UserDTO {
	return UserDTO{
		ID:         user.ID().String(),
		FirstName:  user.FirstName(),
		LastName:   user.LastName(),
		Email:      user.Email().String(),
		Title:      user.Title(),
		Role:       user.Role().String(),
		HasImage:   user.Image() != nil,
		IsActive:   user.IsActive(),
		IsVerified: user.IsVerified(),
		CreatedAt:  user.CreatedAt(),
		LastSignIn: user.LastSignInAt(),
	}
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
