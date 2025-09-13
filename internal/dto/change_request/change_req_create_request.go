package changerequest

type ChangeReqCreateRequest struct {
	Status  string `json:"status"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
