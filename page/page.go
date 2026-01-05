package page

type Spec struct {
	Offset    int
	Limit     int
	SortBy    string
	SortOrder string
}

type Page[T any] struct {
	Total int64
	Data  []T
}
