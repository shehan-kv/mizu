package common

// A generic struct for pagination
type Page[T any] struct {
	Count int64 `json:"count"`
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Data  T     `json:"data"`
}
