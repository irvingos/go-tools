package page

import (
	"testing"
)

type Window struct {
	Offset int
	Limit  int
}

// 或者更领域化一点：Range / Slice
// type Slice struct { Offset, Limit int }

type OrderDirection uint8

const (
	Asc OrderDirection = iota + 1
	Desc
)

// 排序表达建议为强类型字段，而不是 string
type Ordering[F ~string] struct {
	Field     F
	Direction OrderDirection
}

// 领域侧的分页返回建议用 PageResult / Paged
type PageResult[T any] struct {
	Total int64
	Items []T
}

type ConversationSortField string

const (
	ConversationSortByCreatedAt ConversationSortField = "created_at"
	ConversationSortByUpdatedAt ConversationSortField = "updated_at"
)

type ConversationListSpec struct {
	Window   Window
	Ordering *Ordering[ConversationSortField]
}

func TestType(t *testing.T) {
	_ = &ConversationListSpec{
		Window: Window{
			Offset: 0,
			Limit:  10,
		},
		Ordering: &Ordering[ConversationSortField]{
			Field:     "(select * from xxx)",
			Direction: Asc,
		},
	}
}
