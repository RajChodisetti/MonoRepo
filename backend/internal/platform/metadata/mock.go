package metadata

import (
	"context"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct {
	Records map[string]Record
	GetErr  error
}

func (mock *Mock) Get(ctx context.Context, key string) (Record, error) {
	if mock.GetErr != nil {
		return Record{}, mock.GetErr
	}

	record, ok := mock.Records[key]
	if !ok {
		return Record{}, repository.ErrNotFound
	}

	return record, nil
}

var _ Repository = (*Mock)(nil)
