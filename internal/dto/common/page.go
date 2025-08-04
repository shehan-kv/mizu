package common

// A generic struct for pagination
type Page[T any] struct {
	CurrentPage int64 `json:"currentPage"`
	TotalPages  int64 `json:"totalPages"`
	Limit       int64 `json:"limit"`
	Data        T     `json:"data"`
}
