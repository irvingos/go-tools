package tx

import "context"

type Uow interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
