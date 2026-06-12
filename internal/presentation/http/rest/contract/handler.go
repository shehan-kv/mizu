package contract

import (
	"encoding/json"
	"errors"
	"mizu/internal/application/contract"
	"mizu/internal/application/logger"
	domaincontract "mizu/internal/domain/contract"
	domainproject "mizu/internal/domain/project"
	"mizu/internal/presentation/http/rest/middleware"
	"mizu/internal/presentation/http/rest/page"
	"mizu/internal/presentation/http/rest/query"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
)

type ContractHandler struct {
	contractSrv *contract.Service
	log         logger.Logger
}

func NewContractHandler(contractSrv *contract.Service, log logger.Logger) *ContractHandler {
	return &ContractHandler{contractSrv: contractSrv, log: log}
}

func (h *ContractHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /contracts", authMiddleware(http.HandlerFunc(h.ListOverview)))
	mux.Handle("POST /contracts/{projectID}", authMiddleware(http.HandlerFunc(h.CreateContract)))
	mux.Handle("GET /contracts/{contractID}", authMiddleware(http.HandlerFunc(h.GetContract)))
	mux.Handle("GET /contracts/project/{projectID}", authMiddleware(http.HandlerFunc(h.ListOverviewByProject)))
	mux.Handle("GET /contracts/members/{memberID}", authMiddleware(http.HandlerFunc(h.ListOverviewByMember)))
	mux.Handle("POST /contracts/{contractID}/sign", authMiddleware(http.HandlerFunc(h.Sign)))
	mux.Handle("POST /contracts/{contractID}/reject", authMiddleware(http.HandlerFunc(h.Reject)))
	mux.Handle("PUT /contracts/{contractID}/signatories", authMiddleware(http.HandlerFunc(h.ReplaceSignatories)))

	return mux
}

func (h *ContractHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.contractSrv.CreateContract(r.Context(), contract.CreateContractParams{
		ActorID:      actorID,
		ProjectID:    r.PathValue("projectID"),
		Name:         req.Name,
		Terms:        req.Terms,
		SignatoryIDs: req.SignatoryIDs,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ContractHandler) GetContract(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.contractSrv.GetContract(r.Context(), actorID, r.PathValue("contractID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toContractResponse(result))
}

func (h *ContractHandler) ListOverviewByProject(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.contractSrv.ListOverviewByProject(r.Context(), contract.ListByProjectParams{
		ActorID:   actorID,
		ProjectID: r.PathValue("projectID"),
		Keyword:   query.ExtractString(r, "q"),
		Status:    query.ExtractString(r, "status"),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]ContractOverviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toContractOverviewResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[ContractOverviewResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *ContractHandler) ListOverviewByMember(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.contractSrv.ListOverviewByMember(r.Context(), contract.ListByMemberParams{
		ActorID:  actorID,
		MemberID: r.PathValue("memberID"),
		Keyword:  query.ExtractString(r, "q"),
		Status:   query.ExtractString(r, "status"),
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]ContractOverviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toContractOverviewResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[ContractOverviewResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *ContractHandler) ListOverview(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.contractSrv.ListOverviewByMember(r.Context(), contract.ListByMemberParams{
		ActorID:  actorID,
		MemberID: actorID,
		Keyword:  query.ExtractString(r, "q"),
		Status:   query.ExtractString(r, "status"),
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]ContractOverviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toContractOverviewResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[ContractOverviewResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *ContractHandler) Sign(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.contractSrv.Sign(r.Context(), actorID, r.PathValue("contractID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContractHandler) Reject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.contractSrv.Reject(r.Context(), actorID, r.PathValue("contractID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContractHandler) ReplaceSignatories(w http.ResponseWriter, r *http.Request) {
	var req ReplaceSignatoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.contractSrv.ReplaceSignatories(r.Context(), contract.ReplaceSignatoriesParams{
		ActorID:     actorID,
		ContractID:  r.PathValue("contractID"),
		Signatories: req.Signatories,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContractHandler) writeServiceError(w http.ResponseWriter, method string, path string, err error) {

	switch {

	// 404
	case errors.Is(err, domaincontract.ErrContractNotFound):
		response.WriteError(w, http.StatusNotFound, "contract not found")

	case errors.Is(err, domaincontract.ErrContractSignatoryNotFound):
		response.WriteError(w, http.StatusNotFound, "signatory not found")

	case errors.Is(err, domainproject.ErrProjectNotFound):
		response.WriteError(w, http.StatusNotFound, "project not found")

	// 409
	case errors.Is(err, domaincontract.ErrContractSignRequiresPending):
		response.WriteError(w, http.StatusConflict, "contract must be pending to sign")

	case errors.Is(err, domaincontract.ErrContractRejectRequiresPending):
		response.WriteError(w, http.StatusConflict, "contract must be pending to reject")

	case errors.Is(err, domaincontract.ErrContractSignatoryAlreadyActed):
		response.WriteError(w, http.StatusConflict, "signatory has already signed or rejected")

	case errors.Is(err, domaincontract.ErrContractSignatoriesLocked):
		response.WriteError(w, http.StatusConflict, "signatories cannot be replaced after signing has begun")

	// 400
	case errors.Is(err, domaincontract.ErrContractMustHaveAtLeastTwoSignatories):
		response.WriteError(w, http.StatusBadRequest, "contract must have at least two signatories")

	case errors.Is(err, domaincontract.ErrContractSignatoriesMustBeUnique):
		response.WriteError(w, http.StatusBadRequest, "signatories must be unique")

	case errors.Is(err, domaincontract.ErrContractMustHaveClientSignatory):
		response.WriteError(w, http.StatusBadRequest, "contract must have at least one client signatory")

	case errors.Is(err, domaincontract.ErrContractMustHaveTeamSignatory):
		response.WriteError(w, http.StatusBadRequest, "contract must have at least one team signatory")

	// 403
	case errors.Is(err, domainproject.ErrNotProjectMember):
		response.WriteError(w, http.StatusForbidden, "not a project member")

	default:
		h.log.Error(
			"internal server error",
			"method", method,
			"path", path,
			"error", err,
		)

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func toSignatoryResponse(s contract.SignatoryDTO) SignatoryResponse {
	return SignatoryResponse{
		ID:        s.ID,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		Title:     s.Title,
		Role:      s.Role,
		Image:     s.Image,
		Status:    s.Status,
		UpdatedAt: s.UpdatedAt,
	}
}

func toContractResponse(c *contract.ContractDTO) ContractResponse {
	signatories := make([]SignatoryResponse, len(c.Signatories))
	for i, s := range c.Signatories {
		signatories[i] = toSignatoryResponse(s)
	}
	return ContractResponse{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		Name:        c.Name,
		Status:      c.Status,
		Terms:       c.Terms,
		Signatories: signatories,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toContractOverviewResponse(c *contract.ContractOverviewDTO) ContractOverviewResponse {
	signatories := make([]SignatoryResponse, len(c.Signatories))
	for i, s := range c.Signatories {
		signatories[i] = toSignatoryResponse(s)
	}
	return ContractOverviewResponse{
		ID:                    c.ID,
		ProjectID:             c.ProjectID,
		Name:                  c.Name,
		Status:                c.Status,
		MemberSignatoryStatus: c.MemberSignatoryStatus,
		Signatories:           signatories,
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
	}
}
