package shared

type Collection[T any] struct {
	Items      []T
	TotalCount int
}
