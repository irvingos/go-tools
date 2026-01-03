package page

type QuerySpec struct {
	Offset int
	Limit  int
	SortBy string
	Order  string
}

type Page[T any] struct {
	Total int64
	Data  []T
}
