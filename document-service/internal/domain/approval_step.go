package domain

import (
	"time"

	"github.com/google/uuid"
)

type ApprovalStep struct {
	ID              uuid.UUID
	RequestID       uuid.UUID
	StepNumber      int
	ApproverRole    string
	ApproverID      *uuid.UUID
	Status          StepStatus
	DecidedBy       *uuid.UUID
	DecisionComment string
	DecidedAt       *time.Time
	CreatedAt       time.Time
}
