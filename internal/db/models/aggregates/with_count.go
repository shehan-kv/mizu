package aggregates

type WithCount[T any] struct {
	Total int64
	Items []T
}
