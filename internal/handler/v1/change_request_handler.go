package v1

import "mizu/internal/service"

// ChangeRequestHandler provides HTTP handlers
// for change-request related endpoints.
type ChangeRequestHandler struct {
	chngReqSrv *service.ChangeRequestService
}

// NewChangeRequestHandler constructs a new ChangeRequestHandler.
func NewChangeRequestHandler(chngReqSrv *service.ChangeRequestService) *ChangeRequestHandler {
	return &ChangeRequestHandler{
		chngReqSrv: chngReqSrv,
	}
}
