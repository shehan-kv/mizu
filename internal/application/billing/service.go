package billing

import (
	"context"
	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/application/shared"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"time"
)

type Service struct {
	iamRepo     iam.Repository
	billingRepo billing.Repository
	projectRepo project.Repository

	authzSrv *authz.Service

	internalBus eventbus.InternalBus

	idGen  common.IDGenerator
	mailer mailer.Mailer

	logger logger.Logger
}

func NewService(
	iamRepo iam.Repository,
	billingRepo billing.Repository,
	projectRepo project.Repository,
	authzSrv *authz.Service,
	internalBus eventbus.InternalBus,
	idGen common.IDGenerator,
	mailer mailer.Mailer,
	logger logger.Logger,
) *Service {
	return &Service{
		iamRepo:     iamRepo,
		billingRepo: billingRepo,
		projectRepo: projectRepo,
		authzSrv:    authzSrv,
		internalBus: internalBus,
		idGen:       idGen,
		mailer:      mailer,
		logger:      logger,
	}
}

func (s *Service) CreateInvoice(ctx context.Context, params CreateInvoiceParams) error {
	now := time.Now()

	actorID, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actorID); err != nil {
		return err
	}

	code, err := billing.NewCurrencyCode(params.CurrencyCode)
	if err != nil {
		return err
	}

	projectID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, projectID)
	if err != nil {
		return err
	}

	if !p.HasMember(actorID) {
		return project.ErrNotProjectMember
	}

	currency, err := s.billingRepo.GetCurrencyByCode(ctx, code)
	if err != nil {
		return err
	}

	items := make([]billing.Item, len(params.Items))

	for i := range params.Items {
		item, err := newItemFromParams(params.Items[i])
		if err != nil {
			return err
		}
		items[i] = item
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	invID, err := billing.NewInvoiceID(id)
	if err != nil {
		return err
	}

	inv, err := billing.NewInvoice(
		invID,
		projectID,
		params.IsInvoice,
		params.DueDate,
		currency,
		params.Note,
		items,
		now,
	)
	if err != nil {
		return err
	}

	if err := s.billingRepo.Add(ctx, inv); err != nil {
		return err
	}

	s.publishEvents(ctx, inv.PullEvents())

	return nil
}

func newItemFromParams(p CreateItemParams) (billing.Item, error) {
	qty, err := billing.NewQty(p.Qty)
	if err != nil {
		return billing.Item{}, err
	}
	unitPrice, err := billing.NewDecimal(p.UnitPrice)
	if err != nil {
		return billing.Item{}, err
	}
	discountRate, err := billing.NewDecimal(p.DiscountRate)
	if err != nil {
		return billing.Item{}, err
	}
	discountType, err := billing.NewDiscountType(p.DiscountType)
	if err != nil {
		return billing.Item{}, err
	}
	taxRate, err := billing.NewDecimal(p.TaxRate)
	if err != nil {
		return billing.Item{}, err
	}
	taxType, err := billing.NewTaxType(p.TaxType)
	if err != nil {
		return billing.Item{}, err
	}
	return billing.NewItem(p.Description, qty, unitPrice, discountRate, discountType, taxRate, taxType)
}

func (s *Service) ListInvoicesByProject(
	ctx context.Context,
	params ListInvoicesByProjectParams,
) (*shared.Collection[InvoiceOverviewDTO], error) {

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return nil, err
	}

	filter := billing.FilterByProject{
		ProjectID: pID,
		IsInvoice: params.IsInvoice,
		Keyword:   params.Keyword,
	}

	if params.Status != nil {
		status, err := billing.NewStatus(*params.Status)
		if err != nil {
			return nil, err
		}
		filter.Status = &status
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return nil, err
	}

	if !p.HasMember(iam.UserID(params.ActorID)) {
		return nil, project.ErrNotProjectMember
	}

	invs, err := s.billingRepo.ListByProject(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	dtos := make([]InvoiceOverviewDTO, 0, len(invs))

	for i := range invs {
		dtos = append(dtos, InvoiceOverviewDTO{
			ID:            invs[i].ID().String(),
			ProjectID:     p.ID().String(),
			ProjectName:   p.Name().String(),
			IsInvoice:     invs[i].IsInvoice(),
			Status:        invs[i].Status().String(),
			CurrencyCode:  invs[i].Currency().Code().String(),
			Note:          invs[i].Note(),
			TotalTax:      invs[i].TotalTax().String(),
			TotalDiscount: invs[i].TotalDiscount().String(),
			SubTotal:      invs[i].SubTotal().String(),
			DueAt:         invs[i].DueAt(),
			CreatedAt:     invs[i].CreatedAt(),
			UpdatedAt:     invs[i].UpdatedAt(),
		})
	}

	count, err := s.billingRepo.CountByProject(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &shared.Collection[InvoiceOverviewDTO]{Items: dtos, TotalCount: count}, nil
}

func (s *Service) ListInvoicesByMember(
	ctx context.Context,
	params ListInvoicesByMemberParams,
) (*shared.Collection[InvoiceOverviewDTO], error) {

	actorID, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return nil, err
	}

	memberID, err := iam.NewUserID(params.MemberID)
	if err != nil {
		return nil, err
	}

	if err := s.authzSrv.RequireAdministratorOrSelf(ctx, actorID, memberID); err != nil {
		return nil, err
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	filter := billing.FilterByMember{
		MemberID:  memberID,
		IsInvoice: params.IsInvoice,
		Keyword:   params.Keyword,
	}

	if params.Status != nil {
		status, err := billing.NewStatus(*params.Status)
		if err != nil {
			return nil, err
		}
		filter.Status = &status
	}

	invs, err := s.billingRepo.ListByMember(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	projectIDMap := make(map[string]project.ProjectID)

	for i := range invs {
		pID := invs[i].ProjectID()
		projectIDMap[pID.String()] = pID
	}

	projectIDs := make([]project.ProjectID, 0, len(projectIDMap))
	for _, id := range projectIDMap {
		projectIDs = append(projectIDs, id)
	}

	projects, err := s.projectRepo.ListByIDs(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	projectMap := make(map[string]*project.Project, len(projects))

	for i := range projects {
		projectMap[projects[i].ID().String()] = projects[i]
	}

	dtos := make([]InvoiceOverviewDTO, 0, len(invs))

	for i := range invs {
		p := projectMap[invs[i].ProjectID().String()]

		dtos = append(dtos, InvoiceOverviewDTO{
			ID:            invs[i].ID().String(),
			ProjectID:     p.ID().String(),
			ProjectName:   p.Name().String(),
			IsInvoice:     invs[i].IsInvoice(),
			Status:        invs[i].Status().String(),
			CurrencyCode:  invs[i].Currency().Code().String(),
			Note:          invs[i].Note(),
			TotalTax:      invs[i].TotalTax().String(),
			TotalDiscount: invs[i].TotalDiscount().String(),
			SubTotal:      invs[i].SubTotal().String(),
			DueAt:         invs[i].DueAt(),
			CreatedAt:     invs[i].CreatedAt(),
			UpdatedAt:     invs[i].UpdatedAt(),
		})
	}

	count, err := s.billingRepo.CountByMember(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &shared.Collection[InvoiceOverviewDTO]{Items: dtos, TotalCount: count}, nil
}

func (s *Service) GetInvoice(ctx context.Context, actorID string, invoiceID string) (InvoiceDTO, error) {

	invID, err := billing.NewInvoiceID(invoiceID)
	if err != nil {
		return InvoiceDTO{}, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return InvoiceDTO{}, err
	}

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return InvoiceDTO{}, err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return InvoiceDTO{}, err
	}

	if !p.HasMember(actor) {
		return InvoiceDTO{}, project.ErrNotProjectMember
	}

	items := inv.Items()
	itemdto := make([]InvoiceItemDTO, 0, len(items))
	for i := range items {
		itemdto = append(itemdto, InvoiceItemDTO{
			Description:           items[i].Description(),
			Qty:                   items[i].Qty().String(),
			UnitPrice:             items[i].UnitPrice().String(),
			DiscountRate:          items[i].DiscountRate().String(),
			DiscountType:          items[i].DiscountType().String(),
			TaxRate:               items[i].TaxRate().String(),
			TaxType:               items[i].TaxType().String(),
			DiscountAmountPerUnit: items[i].DiscountAmountPerUnit().String(),
			TaxableBasePerUnit:    items[i].TaxableBasePerUnit().String(),
			TaxAmountPerUnit:      items[i].TaxAmountPerUnit().String(),
			LineGross:             items[i].LineGross().String(),
			LineDiscount:          items[i].LineDiscount().String(),
			LineNet:               items[i].LineNet().String(),
			LineTax:               items[i].LineTax().String(),
			LineTotal:             items[i].LineTotal().String(),
		})
	}

	return InvoiceDTO{
		ID:            inv.ID().String(),
		ProjectID:     inv.ProjectID().String(),
		ProjectName:   p.Name().String(),
		IsInvoice:     inv.IsInvoice(),
		Status:        inv.Status().String(),
		DueAt:         inv.DueAt(),
		CurrencyCode:  inv.Currency().Code().String(),
		Note:          inv.Note(),
		CreatedAt:     inv.CreatedAt(),
		UpdatedAt:     inv.UpdatedAt(),
		TotalTax:      inv.TotalTax().String(),
		TotalDiscount: inv.TotalDiscount().String(),
		SubTotal:      inv.SubTotal().String(),
		Items:         itemdto,
	}, nil
}

func (s *Service) EmailInvoice(ctx context.Context, invoiceID string, actorID string, memberID string) error {

	invID, err := billing.NewInvoiceID(invoiceID)
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

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorStaffOrSelf(ctx, actor, member); err != nil {
		return err
	}

	m, err := s.iamRepo.GetByID(ctx, member)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) || !p.HasMember(member) {
		return project.ErrNotProjectMember
	}

	items := inv.Items()
	emailItems := make([]mailer.InvoiceItem, 0, len(items))
	for _, item := range items {
		emailItems = append(emailItems, mailer.InvoiceItem{
			Description: item.Description(),

			Qty:       item.Qty().String(),
			UnitPrice: item.UnitPrice().String(),

			DiscountRate: item.DiscountRate().String(),
			DiscountType: item.DiscountType().String(),

			TaxRate: item.TaxRate().String(),
			TaxType: item.TaxType().String(),

			LineGross:    item.LineGross().String(),
			LineDiscount: item.LineDiscount().String(),
			LineNet:      item.LineNet().String(),
			LineTax:      item.LineTax().String(),
			LineTotal:    item.LineTotal().String(),
		})
	}

	id := inv.ID().String()
	suffix := id[max(0, len(id)-8):]

	subjectType := "Quote"
	if inv.IsInvoice() {
		subjectType = "Invoice"
	}

	subject := subjectType + ": #" + suffix

	email := mailer.InvoiceEmail{
		Subject:        subject,
		RecipientEmail: m.Email().String(),

		InvoiceID: inv.ID().String(),
		ProjectID: inv.ProjectID().String(),

		Status:       inv.Status().String(),
		CurrencyName: inv.Currency().Name().String(),
		CurrencyCode: inv.Currency().Code().String(),

		DueAt: inv.DueAt(),
		Note:  inv.Note(),

		Items: emailItems,

		SubTotal:      inv.SubTotal().String(),
		TotalTax:      inv.TotalTax().String(),
		TotalDiscount: inv.TotalDiscount().String(),
	}

	return s.mailer.SendInvoiceEmail(ctx, email)
}

func (s *Service) AcceptInvoice(ctx context.Context, actorID string, invoiceID string) error {

	now := time.Now()

	invID, err := billing.NewInvoiceID(invoiceID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireClient(ctx, actor); err != nil {
		return err
	}

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if err := inv.Accept(now); err != nil {
		return err
	}

	if err := s.billingRepo.Save(ctx, inv); err != nil {
		return err
	}

	s.publishEvents(ctx, inv.PullEvents())

	return nil
}

func (s *Service) RejectInvoice(ctx context.Context, actorID string, invoiceID string) error {

	now := time.Now()

	invID, err := billing.NewInvoiceID(invoiceID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireClient(ctx, actor); err != nil {
		return err
	}

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if err := inv.Reject(now); err != nil {
		return err
	}

	if err := s.billingRepo.Save(ctx, inv); err != nil {
		return err
	}

	s.publishEvents(ctx, inv.PullEvents())

	return nil
}

func (s *Service) PayInvoice(ctx context.Context, actorID string, invoiceID string) error {

	now := time.Now()

	invID, err := billing.NewInvoiceID(invoiceID)
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

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if err := inv.Pay(now); err != nil {
		return err
	}

	if err := s.billingRepo.Save(ctx, inv); err != nil {
		return err
	}

	s.publishEvents(ctx, inv.PullEvents())

	return nil
}

func (s *Service) CancelInvoice(ctx context.Context, actorID string, invoiceID string) error {

	now := time.Now()

	invID, err := billing.NewInvoiceID(invoiceID)
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

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if err := inv.Cancel(now); err != nil {
		return err
	}

	if err := s.billingRepo.Save(ctx, inv); err != nil {
		return err
	}

	s.publishEvents(ctx, inv.PullEvents())

	return nil
}

func (s *Service) ConvertToInvoice(ctx context.Context, actorID string, invoiceID string) error {
	now := time.Now()

	invID, err := billing.NewInvoiceID(invoiceID)
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

	inv, err := s.billingRepo.Get(ctx, invID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, inv.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if err := inv.ConvertToInvoice(now); err != nil {
		return err
	}

	if err := s.billingRepo.Save(ctx, inv); err != nil {
		return err
	}

	s.publishEvents(ctx, inv.PullEvents())

	return nil
}

func (s *Service) ListPaidCountByProject(ctx context.Context, actorID string, projectID string) ([]MetricDTO, error) {

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
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

	m, err := s.billingRepo.ListMonthlyPaidCountByProject(ctx, pID)
	if err != nil {
		return nil, err
	}

	dto := make([]MetricDTO, 0, len(m))
	for i := range m {
		dto = append(dto, MetricDTO{
			Key:   m[i].Key(),
			Value: m[i].Value(),
		})
	}

	return dto, nil
}

func (s *Service) ListPaidCountByMember(ctx context.Context, actorID string, memberID string) ([]MetricDTO, error) {

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

	m, err := s.billingRepo.ListMonthlyPaidCountByMember(ctx, memID)
	if err != nil {
		return nil, err
	}

	dto := make([]MetricDTO, 0, len(m))
	for i := range m {
		dto = append(dto, MetricDTO{
			Key:   m[i].Key(),
			Value: m[i].Value(),
		})
	}

	return dto, nil
}

func (s *Service) GetBillingSummaryByMember(ctx context.Context, actorID string, memberID string) (BillingSummaryDTO, error) {

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return BillingSummaryDTO{}, err
	}

	memID, err := iam.NewUserID(memberID)
	if err != nil {
		return BillingSummaryDTO{}, err
	}

	if err := s.authzSrv.RequireAdministratorOrSelf(ctx, actor, memID); err != nil {
		return BillingSummaryDTO{}, err
	}

	overview, err := s.billingRepo.GetBillingOverviewByMember(ctx, memID)
	if err != nil {
		return BillingSummaryDTO{}, err
	}

	toDTO := func(metrics []billing.BillingOverviewMetric) []BillingSummaryMetricDTO {
		dtos := make([]BillingSummaryMetricDTO, 0, len(metrics))
		for i := range metrics {
			dtos = append(dtos, BillingSummaryMetricDTO{
				CurrencyCode: metrics[i].CurrencyCode().String(),
				Amount:       metrics[i].Amount().String(),
				Count:        metrics[i].Count(),
			})
		}
		return dtos
	}

	dto := BillingSummaryDTO{
		InvoicesPaid:      toDTO(overview.InvoicesPaid()),
		InvoicesPending:   toDTO(overview.InvoicesPending()),
		InvoicesAccepted:  toDTO(overview.InvoicesAccepted()),
		InvoicesRejected:  toDTO(overview.InvoicesRejected()),
		InvoicesCancelled: toDTO(overview.InvoicesCancelled()),
		QuotesPending:     toDTO(overview.QuotesPending()),
		QuotesRejected:    toDTO(overview.QuotesRejected()),
	}

	return dto, nil
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
