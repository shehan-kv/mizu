package message

import (
	"encoding/json"
	"errors"
	"io"
	"mizu/internal/application/logger"
	"mizu/internal/application/message"
	domainmessage "mizu/internal/domain/message"
	domainproject "mizu/internal/domain/project"
	"mizu/internal/presentation/http/rest/middleware"
	"mizu/internal/presentation/http/rest/page"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
	"strconv"
)

type MessageHandler struct {
	msgSrv *message.Service
	log    logger.Logger

	maxFileSizeMB int64
}

func NewMessageHandler(msgSrv *message.Service, log logger.Logger, maxFileSizeMB int64) *MessageHandler {
	return &MessageHandler{msgSrv: msgSrv, log: log, maxFileSizeMB: maxFileSizeMB}
}

func (h *MessageHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /messages/channels", authMiddleware(http.HandlerFunc(h.CreateChannel)))
	mux.Handle("POST /messages/{channelID}", authMiddleware(http.HandlerFunc(h.CreateMessage)))
	mux.Handle("GET /messages/files/{fileID}", authMiddleware(http.HandlerFunc(h.DownloadFile)))
	mux.Handle("GET /messages/channels", authMiddleware(http.HandlerFunc(h.ListChannels)))
	mux.Handle("GET /messages/members/{memberID}/channels", authMiddleware(http.HandlerFunc(h.ListChannelsByMember)))
	mux.Handle("GET /messages/channels/{channelID}/members", authMiddleware(http.HandlerFunc(h.ListChannelMembers)))
	mux.Handle("PUT /messages/channels/{channelID}/members", authMiddleware(http.HandlerFunc(h.ReplaceChannelMembers)))
	mux.Handle("GET /messages/channels/{channelID}", authMiddleware(http.HandlerFunc(h.ListChannelMessages)))
	mux.Handle("POST /messages/channels/{channelID}/files", authMiddleware(http.HandlerFunc(h.UploadFile)))
	mux.Handle("GET /messages/channels/{channelID}/files", authMiddleware(http.HandlerFunc(h.ListChannelFiles)))
	mux.Handle("GET /messages/projects/{projectID}/files", authMiddleware(http.HandlerFunc(h.ListProjectFiles)))

	return mux
}

func (h *MessageHandler) ReplaceChannelMembers(w http.ResponseWriter, r *http.Request) {
	var req ReplaceChannelMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.msgSrv.ReplaceChannelMembers(
		r.Context(),
		r.PathValue("channelID"),
		req.MemberIDs, actorID,
	); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MessageHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.msgSrv.CreateChannel(r.Context(), message.CreateChannelParams{
		ActorID:   actorID,
		ProjectID: req.ProjectID,
		Name:      req.Name,
		MemberIDs: req.MemberIDs,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *MessageHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.msgSrv.ListChannelsByMember(r.Context(), actorID, actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	channels := make([]ChannelResponse, len(result))
	for i := range result {
		channels[i] = toChannelResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, channels)
}

func (h *MessageHandler) ListChannelsByMember(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.msgSrv.ListChannelsByMember(r.Context(), actorID, r.PathValue("memberID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	channels := make([]ChannelResponse, len(result))
	for i := range result {
		channels[i] = toChannelResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, channels)
}

func (h *MessageHandler) ListChannelMembers(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.msgSrv.ListChannelMembers(r.Context(), actorID, r.PathValue("channelID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	members := make([]MemberResponse, len(result))
	for i := range result {
		members[i] = toMemberResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, members)
}

func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.msgSrv.CreateMessage(r.Context(), message.CreateMessageParams{
		ActorID:   actorID,
		ChannelID: r.PathValue("channelID"),
		Content:   req.Content,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *MessageHandler) ListChannelMessages(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.msgSrv.ListChannelMessages(r.Context(), message.ListChannelMessagesParams{
		ActorID:   actorID,
		ChannelID: r.PathValue("channelID"),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	messages := make([]MessageResponse, len(result.Items))
	for i := range result.Items {
		messages[i] = toMessageResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[MessageResponse]{
		Items:      messages,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *MessageHandler) UploadFile(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	// HARD LIMIT: reject anything over the specified size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxFileSizeMB<<20)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "file too large or invalid form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "file is required")
		return
	}

	defer file.Close()

	err = h.msgSrv.UploadFile(r.Context(), message.UploadFileParams{
		ActorID:   actorID,
		ChannelID: r.PathValue("channelID"),
		FileName:  header.Filename,
		MimeType:  header.Header.Get("Content-Type"),
		Size:      header.Size,
		Reader:    file,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *MessageHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.msgSrv.DownloadFile(r.Context(), actorID, r.PathValue("fileID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	defer result.Reader.Close()

	w.Header().Set("Content-Type", result.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(result.Name))
	w.Header().Set("Content-Length", strconv.FormatInt(result.Size, 10))

	_, err = io.Copy(w, result.Reader)
	if err != nil {
		return
	}
}

func (h *MessageHandler) ListChannelFiles(w http.ResponseWriter, r *http.Request) {

	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	keyword := r.URL.Query().Get("q")
	var kw *string
	if keyword != "" {
		kw = &keyword
	}

	result, err := h.msgSrv.ListChannelFiles(
		r.Context(),
		message.ListChannelFilesParams{
			ActorID:   actorID,
			ChannelID: r.PathValue("channelID"),
			Keyword:   kw,
			Limit:     p.Limit,
			Offset:    p.Offset,
		},
	)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	files := make([]FileResponse, len(result.Items))
	for i := range result.Items {
		files[i] = toFileResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[FileResponse]{
		Items:      files,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *MessageHandler) ListProjectFiles(w http.ResponseWriter, r *http.Request) {

	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	keyword := r.URL.Query().Get("q")
	var kw *string
	if keyword != "" {
		kw = &keyword
	}

	result, err := h.msgSrv.ListProjectFiles(
		r.Context(),
		message.ListProjectFilesParams{
			ActorID:   actorID,
			ProjectID: r.PathValue("projectID"),
			Keyword:   kw,
			Limit:     p.Limit,
			Offset:    p.Offset,
		},
	)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	files := make([]FileResponse, len(result.Items))
	for i := range result.Items {
		files[i] = toFileResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[FileResponse]{
		Items:      files,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *MessageHandler) writeServiceError(w http.ResponseWriter, method string, path string, err error) {

	switch {

	// 404
	case errors.Is(err, domainmessage.ErrChannelNotFound):
		response.WriteError(w, http.StatusNotFound, "channel not found")

	case errors.Is(err, domainmessage.ErrChannelMemberNotFound):
		response.WriteError(w, http.StatusNotFound, "channel member not found")

	case errors.Is(err, domainmessage.ErrMessageNotFound):
		response.WriteError(w, http.StatusNotFound, "message not found")

	// 409
	case errors.Is(err, domainmessage.ErrChannelMemberAlreadyExists):
		response.WriteError(w, http.StatusConflict, "member already exists in channel")

	case errors.Is(err, domainmessage.ErrChannelConcurrentModification):
		response.WriteError(w, http.StatusConflict, "channel was modified by another request")

	// 400
	case errors.Is(err, domainmessage.ErrChannelMustHaveAtLeastTwoMembers):
		response.WriteError(w, http.StatusBadRequest, "channel must have at least two members")

	case errors.Is(err, domainmessage.ErrChannelMembersMustBeUnique):
		response.WriteError(w, http.StatusBadRequest, "channel members must be unique")

	case errors.Is(err, domainmessage.ErrChannelMustHaveStaffMember):
		response.WriteError(w, http.StatusBadRequest, "channel must have at least one staff member")

	case errors.Is(err, domainmessage.ErrChannelNameCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "channel name cannot be empty")

	case errors.Is(err, domainmessage.ErrChannelIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "channel id cannot be empty")

	case errors.Is(err, domainmessage.ErrMessageContentCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "message content cannot be empty")

	case errors.Is(err, domainmessage.ErrMessageContentTooLong):
		response.WriteError(w, http.StatusBadRequest, "message content is too long")

	case errors.Is(err, domainmessage.ErrMessageInvalidSender):
		response.WriteError(w, http.StatusBadRequest, "invalid message sender")

	case errors.Is(err, domainmessage.ErrMessageIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "message id cannot be empty")

	// 403
	case errors.Is(err, domainmessage.ErrNotChannelMember):
		response.WriteError(w, http.StatusForbidden, "not a channel member")

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

func toChannelResponse(c *message.ChannelDTO) ChannelResponse {
	return ChannelResponse{
		ID:        c.ID,
		ProjectID: c.ProjectID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toMemberResponse(m *message.MemberDTO) MemberResponse {
	return MemberResponse{
		ID:        m.ID,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		HasImage:  m.HasImage,
		Title:     m.Title,
		Role:      m.Role,
	}
}

func toMessageSenderResponse(s *message.MessageSenderDTO) *MessageSenderResponse {
	if s == nil {
		return nil
	}
	return &MessageSenderResponse{
		ID:        s.ID,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		Image:     s.Image,
		Title:     s.Title,
		Role:      s.Role,
	}
}

func toMessageResponse(m *message.MessageDTO) MessageResponse {
	return MessageResponse{
		ID:        m.ID,
		ChannelID: m.ChannelID,
		IsSystem:  m.IsSystem,
		Content:   m.Content,
		Sender:    toMessageSenderResponse(m.Sender),
		CreatedAt: m.CreatedAt,
	}
}

func toFileResponse(f *message.FileDTO) FileResponse {
	return FileResponse{
		ID:         f.ID,
		ChannelID:  f.ChannelID,
		User:       *toMessageSenderResponse(&f.User),
		Name:       f.Name,
		MimeType:   f.MimeType,
		Size:       f.Size,
		UploadedAt: f.UploadedAt,
	}
}
