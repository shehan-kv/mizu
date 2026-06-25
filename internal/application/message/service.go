package message

import (
	"context"
	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/filestore"
	"mizu/internal/application/logger"
	"mizu/internal/application/message/integration"
	"mizu/internal/application/shared"
	"mizu/internal/application/uow"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"slices"
	"time"
)

type Service struct {
	iamRepo     iam.Repository
	channelRepo message.ChannelRepository
	messageRepo message.MessageRepository
	fileRepo    message.FileRepository
	projectRepo project.Repository

	idGen common.IDGenerator

	internalBus eventbus.InternalBus
	externalBus eventbus.ExternalBus

	authzSrv   *authz.Service
	messageSrv *message.Service

	fileStore filestore.Store

	uow uow.UnitOfWork

	log logger.Logger
}

func NewService(
	iamRepo iam.Repository,
	channelRepo message.ChannelRepository,
	messageRepo message.MessageRepository,
	fileRepo message.FileRepository,
	projectRepo project.Repository,
	idGen common.IDGenerator,
	internalBus eventbus.InternalBus,
	externalBus eventbus.ExternalBus,
	authzSrv *authz.Service,
	messageSrv *message.Service,
	fileStore filestore.Store,
	uow uow.UnitOfWork,
	log logger.Logger,
) *Service {

	return &Service{
		iamRepo:     iamRepo,
		channelRepo: channelRepo,
		messageRepo: messageRepo,
		fileRepo:    fileRepo,
		projectRepo: projectRepo,
		idGen:       idGen,
		internalBus: internalBus,
		externalBus: externalBus,
		authzSrv:    authzSrv,
		messageSrv:  messageSrv,
		fileStore:   fileStore,
		uow:         uow,
		log:         log,
	}
}

func (s *Service) ReplaceChannelMembers(ctx context.Context, channelID string, memberIDs []string, actorID string) error {
	now := time.Now()

	cID, err := message.NewChannelID(channelID)
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

	c, err := s.channelRepo.Get(ctx, cID)
	if err != nil {
		return err
	}

	if !c.HasMember(actor) {
		return message.ErrNotChannelMember
	}

	ids := make([]iam.UserID, len(memberIDs))
	for i := range memberIDs {
		id, err := iam.NewUserID(memberIDs[i])
		if err != nil {
			return err
		}

		ids[i] = id
	}

	if !slices.Contains(ids, actor) {
		ids = append(ids, actor)
	}

	allExist, err := s.iamRepo.ExistsAll(ctx, ids)
	if err != nil {
		return err
	}
	if !allExist {
		return iam.ErrUserNotFound
	}

	c.ReplaceMembers(ids, now)

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.channelRepo.Save(ctx, c)
	})
}

func (s *Service) CreateChannel(ctx context.Context, params CreateChannelParams) error {

	now := time.Now()

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	name, err := message.NewChannelName(params.Name)
	if err != nil {
		return err
	}

	memberIDs := make([]iam.UserID, len(params.MemberIDs))
	for i := range params.MemberIDs {
		memID, err := iam.NewUserID(params.MemberIDs[i])
		if err != nil {
			return err
		}

		memberIDs[i] = memID
	}

	if !slices.Contains(memberIDs, actor) {
		memberIDs = append(memberIDs, actor)
	}

	var projectID *project.ProjectID
	if params.ProjectID != nil {
		pID, err := project.NewProjectID(*params.ProjectID)
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
		projectID = &pID

		exists, err := s.iamRepo.ExistsAll(ctx, memberIDs)
		if err != nil {
			return err
		}
		if !exists {
			return message.ErrNotChannelMember
		}
	}

	if params.ProjectID == nil {
		users, err := s.iamRepo.ListByIDs(ctx, memberIDs, iam.UserFilter{})
		if err != nil {
			return err
		}
		if err := s.messageSrv.ValidateStandaloneChannelMembers(users); err != nil {
			return err
		}
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	chID, err := message.NewChannelID(id)
	if err != nil {
		return err
	}

	ch, err := message.NewChannel(
		chID,
		projectID,
		name,
		memberIDs,
		now,
	)
	if err != nil {
		return err
	}

	return s.channelRepo.Add(ctx, ch)
}

func (s *Service) ListChannelsByMember(ctx context.Context, actorID string, memberID string) ([]ChannelDTO, error) {

	memID, err := iam.NewUserID(memberID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	if err := s.authzSrv.RequireAdministratorOrSelf(ctx, actor, memID); err != nil {
		return nil, err
	}

	c, err := s.channelRepo.ListByMember(ctx, memID)
	if err != nil {
		return nil, err
	}

	cdto := make([]ChannelDTO, len(c))
	for i := range c {

		var projectID *string
		if pid := c[i].ProjectID(); pid != nil {
			s := pid.String()
			projectID = &s
		}

		cdto[i] = ChannelDTO{
			ID:        c[i].ID().String(),
			ProjectID: projectID,
			Name:      c[i].Name().String(),
			CreatedAt: c[i].CreatedAt(),
			UpdatedAt: c[i].UpdatedAt(),
		}
	}

	return cdto, nil
}

func (s *Service) ListChannelMembers(ctx context.Context, actorID string, channelID string) ([]MemberDTO, error) {

	cID, err := message.NewChannelID(channelID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	c, err := s.channelRepo.Get(ctx, cID)
	if err != nil {
		return nil, err
	}

	if !c.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	m, err := s.iamRepo.ListByIDs(ctx, c.Members(), iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	mdto := make([]MemberDTO, len(m))
	for i := range m {
		mdto[i] = MemberDTO{
			ID:        m[i].ID().String(),
			FirstName: m[i].FirstName(),
			LastName:  m[i].LastName(),
			HasImage:  m[i].Image() != nil,
			Title:     m[i].Title(),
			Role:      m[i].Role().String(),
		}
	}

	return mdto, nil

}

func (s *Service) CreateMessage(ctx context.Context, params CreateMessageParams) error {

	now := time.Now()

	cID, err := message.NewChannelID(params.ChannelID)
	if err != nil {
		return err
	}

	c, err := s.channelRepo.Get(ctx, cID)
	if err != nil {
		return err
	}

	actorID, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	actor, err := s.iamRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}

	if !c.HasMember(actorID) {
		return message.ErrNotChannelMember
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	msgID, err := message.NewMessageID(id)
	if err != nil {
		return err
	}

	content, err := message.NewContent(params.Content)
	if err != nil {
		return err
	}

	m, err := message.NewUserMessage(msgID, cID, actorID, content, now)
	if err != nil {
		return err
	}

	if err := s.messageRepo.Add(ctx, m); err != nil {
		return err
	}

	members := c.Members()
	to := make([]string, 0, len(members))
	for i := range members {
		to = append(to, members[i].String())
	}

	firstName := actor.FirstName()
	lastName := actor.LastName()
	role := actor.Role().String()

	var image *string
	if actor.Image() != nil {
		name := actor.Image().Name().String()
		image = &name
	}

	if err := s.externalBus.Publish(ctx, integration.MessageBroadcast{
		MessageID:       msgID.String(),
		ChannelID:       m.ChannelID().String(),
		To:              to,
		SenderID:        actorID.String(),
		SenderFirstName: &firstName,
		SenderLastName:  &lastName,
		SenderImage:     image,
		SenderTitle:     actor.Title(),
		SenderRole:      &role,
		IsSystem:        false,
		Content:         m.Content().String(),
		OccurredAt:      m.CreatedAt(),
	}); err != nil {

		s.log.Error("failed to publish integration event",
			"event_type", integration.EventMessageBroadcast,
			"err", err)
	}

	return nil
}

func (s *Service) ListChannelMessages(ctx context.Context, params ListChannelMessagesParams) (shared.Collection[MessageDTO], error) {

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}

	channelID, err := message.NewChannelID(params.ChannelID)
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}

	channel, err := s.channelRepo.Get(ctx, channelID)
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}
	if !channel.HasMember(actor) {
		return shared.Collection[MessageDTO]{}, message.ErrNotChannelMember
	}

	messages, err := s.messageRepo.ListByChannel(ctx, channelID, page)
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}

	seen := make(map[iam.UserID]struct{})
	senderIDs := make([]iam.UserID, 0)
	for _, m := range messages {
		if m.IsSystemMessage() {
			continue
		}
		if _, exists := seen[m.SenderID()]; !exists {
			seen[m.SenderID()] = struct{}{}
			senderIDs = append(senderIDs, m.SenderID())
		}
	}

	users, err := s.iamRepo.ListByIDs(ctx, senderIDs, iam.UserFilter{})
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}

	userMap := make(map[iam.UserID]iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = *users[i]
	}

	dtos := make([]MessageDTO, len(messages))
	for i := range messages {
		dto := MessageDTO{
			ID:        messages[i].ID().String(),
			ChannelID: messages[i].ChannelID().String(),
			IsSystem:  messages[i].IsSystemMessage(),
			Content:   messages[i].Content().String(),
			CreatedAt: messages[i].CreatedAt(),
		}
		if !messages[i].IsSystemMessage() {
			if u, ok := userMap[messages[i].SenderID()]; ok {
				var image *string
				if u.Image() != nil {
					name := u.Image().Name().String()
					image = &name
				}

				dto.Sender = &MessageSenderDTO{
					ID:        u.ID().String(),
					FirstName: u.FirstName(),
					LastName:  u.LastName(),
					Image:     image,
					Title:     u.Title(),
					Role:      u.Role().String(),
				}
			}
		}
		dtos[i] = dto
	}

	count, err := s.messageRepo.CountByChannel(ctx, channelID)
	if err != nil {
		return shared.Collection[MessageDTO]{}, err
	}

	return shared.Collection[MessageDTO]{Items: dtos, TotalCount: count}, nil
}

func (s *Service) UploadFile(ctx context.Context, params UploadFileParams) error {

	now := time.Now()

	cID, err := message.NewChannelID(params.ChannelID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	ch, err := s.channelRepo.Get(ctx, cID)
	if err != nil {
		return err
	}
	if !ch.HasMember(actor) {
		return message.ErrNotChannelMember
	}

	fileIDString, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	fID, err := message.NewFileID(fileIDString)
	if err != nil {
		return err
	}

	if err := s.fileStore.Save(ctx, fileIDString, params.Reader); err != nil {
		return err
	}

	originalName, err := message.NewFileName(params.FileName)
	if err != nil {
		return err
	}

	savedName, err := message.NewFileName(fileIDString)
	if err != nil {
		return err
	}

	file := message.NewFile(
		fID,
		cID,
		actor,
		originalName,
		savedName,
		fileIDString,
		params.MimeType,
		params.Size,
		now,
	)

	if err := s.fileRepo.Add(ctx, file); err != nil {
		return err
	}

	s.publishEvents(ctx, file.PullEvents())

	return nil
}

func (s *Service) DownloadFile(ctx context.Context, actorID string, fileID string) (*FileDownloadDTO, error) {

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	fID, err := message.NewFileID(fileID)
	if err != nil {
		return nil, err
	}

	file, err := s.fileRepo.Get(ctx, fID)
	if err != nil {
		return nil, err
	}

	channel, err := s.channelRepo.Get(ctx, file.ChannelID())
	if err != nil {
		return nil, err
	}

	if !channel.HasMember(actor) {
		return nil, message.ErrNotChannelMember
	}

	reader, err := s.fileStore.Open(ctx, file.StorageKey())
	if err != nil {
		return nil, err
	}

	return &FileDownloadDTO{
		Name:       file.OriginalName().String(),
		MimeType:   file.MimeType(),
		Size:       file.Size(),
		Reader:     reader,
		UploadedAt: file.UploadedAt(),
	}, nil
}

func (s *Service) ListChannelFiles(ctx context.Context, params ListChannelFilesParams) (shared.Collection[FileDTO], error) {
	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	channelID, err := message.NewChannelID(params.ChannelID)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	filter := message.ChannelFileFilter{
		ChannelID: channelID,
		Keyword:   params.Keyword,
	}

	channel, err := s.channelRepo.Get(ctx, channelID)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}
	if !channel.HasMember(actor) {
		return shared.Collection[FileDTO]{}, message.ErrNotChannelMember
	}

	files, err := s.fileRepo.ListByChannel(ctx, filter, page)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	seen := make(map[iam.UserID]struct{})
	userIDs := make([]iam.UserID, 0, len(files))

	for i := range files {
		uid := files[i].UserID()
		if _, ok := seen[uid]; !ok {
			seen[uid] = struct{}{}
			userIDs = append(userIDs, uid)
		}
	}

	users, err := s.iamRepo.ListByIDs(ctx, userIDs, iam.UserFilter{})
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	userMap := make(map[iam.UserID]iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = *users[i]
	}

	dtos := make([]FileDTO, len(files))
	for i := range files {
		dto := FileDTO{
			ID:         files[i].ID().String(),
			ChannelID:  files[i].ChannelID().String(),
			Name:       files[i].OriginalName().String(),
			MimeType:   files[i].MimeType(),
			Size:       files[i].Size(),
			UploadedAt: files[i].UploadedAt(),
		}

		dto.User = MessageSenderDTO{}

		if u, ok := userMap[files[i].UserID()]; ok {
			var image *string
			if u.Image() != nil {
				name := u.Image().Name().String()
				image = &name
			}

			dto.User.ID = u.ID().String()
			dto.User.FirstName = u.FirstName()
			dto.User.LastName = u.LastName()
			dto.User.Image = image
			dto.User.Title = u.Title()
			dto.User.Role = u.Role().String()
		}
		dtos[i] = dto
	}

	count, err := s.fileRepo.CountByChannel(ctx, message.ChannelFileFilter{
		ChannelID: channelID,
		Keyword:   params.Keyword,
	})
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	return shared.Collection[FileDTO]{
		Items:      dtos,
		TotalCount: count,
	}, nil

}

func (s *Service) ListProjectFiles(ctx context.Context, params ListProjectFilesParams) (shared.Collection[FileDTO], error) {
	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	filter := message.ProjectFileFilter{
		MemberID:  actor,
		ProjectID: pID,
		Keyword:   params.Keyword,
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}
	if !p.HasMember(actor) {
		return shared.Collection[FileDTO]{}, project.ErrNotProjectMember
	}

	files, err := s.fileRepo.ListByProject(ctx, filter, page)
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	seen := make(map[iam.UserID]struct{})
	userIDs := make([]iam.UserID, 0, len(files))

	for i := range files {
		uid := files[i].UserID()
		if _, ok := seen[uid]; !ok {
			seen[uid] = struct{}{}
			userIDs = append(userIDs, uid)
		}
	}

	users, err := s.iamRepo.ListByIDs(ctx, userIDs, iam.UserFilter{})
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	userMap := make(map[iam.UserID]iam.User, len(users))
	for i := range users {
		userMap[users[i].ID()] = *users[i]
	}

	dtos := make([]FileDTO, len(files))
	for i := range files {
		dto := FileDTO{
			ID:         files[i].ID().String(),
			ChannelID:  files[i].ChannelID().String(),
			Name:       files[i].OriginalName().String(),
			MimeType:   files[i].MimeType(),
			Size:       files[i].Size(),
			UploadedAt: files[i].UploadedAt(),
		}

		dto.User = MessageSenderDTO{}

		if u, ok := userMap[files[i].UserID()]; ok {
			var image *string
			if u.Image() != nil {
				name := u.Image().Name().String()
				image = &name
			}

			dto.User.ID = u.ID().String()
			dto.User.FirstName = u.FirstName()
			dto.User.LastName = u.LastName()
			dto.User.Image = image
			dto.User.Title = u.Title()
			dto.User.Role = u.Role().String()
		}
		dtos[i] = dto
	}

	count, err := s.fileRepo.CountByProject(ctx, message.ProjectFileFilter{
		ProjectID: pID,
		Keyword:   params.Keyword,
	})
	if err != nil {
		return shared.Collection[FileDTO]{}, err
	}

	return shared.Collection[FileDTO]{
		Items:      dtos,
		TotalCount: count,
	}, nil

}

func (s *Service) publishEvents(ctx context.Context, events []common.Event) {
	for _, event := range events {
		if err := s.internalBus.Publish(ctx, event); err != nil {

			s.log.Warn("failed to publish event",
				"event_type", event.EventType(),
				"err", err,
			)
		}
	}
}
