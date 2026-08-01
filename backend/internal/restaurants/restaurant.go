package restaurants

import (
	"time"

	"github.com/google/uuid"
)

type Restaurant struct {
	ID                    uuid.UUID
	Name                  string
	Email                 string
	Phone                 string
	Address               string
	OCRStatus             string
	Status                string
	IsContacted           bool
	ShownInterest         bool
	EmailSent             bool
	EmailSendCount        int
	LastEmailSentAt       *time.Time
	LastEmailSendSequence *int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreateInput struct {
	Name  string
	Email string
}

type ListFilter struct {
	Restaurant      string
	Status          string
	IsContacted     *bool
	ShownInterest   *bool
	OCRStatus       string
	IncludeArchived bool
}

type UpdateInput struct {
	Name          *string
	Email         *string
	IsContacted   *bool
	ShownInterest *bool
	Status        *string
}
