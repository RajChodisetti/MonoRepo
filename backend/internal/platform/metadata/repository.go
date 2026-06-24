package metadata

import "context"

type Repository interface {
	Get(ctx context.Context, key string) (Record, error)
}
