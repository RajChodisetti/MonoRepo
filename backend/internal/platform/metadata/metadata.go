package metadata

import (
	"encoding/json"
	"time"
)

type Record struct {
	Key       string
	Value     json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}
