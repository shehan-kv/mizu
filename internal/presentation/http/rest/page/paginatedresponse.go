package page

type PaginatedResponse[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
}
