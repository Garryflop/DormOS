package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkflowDefinition struct {
	ID           uuid.UUID
	DocumentType string
	Name         string
	Description  string
	Steps        []WorkflowStepDef
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type WorkflowStepDef struct {
	StepNumber   int     `json:"step_number"`
	ApproverRole string  `json:"approver_role"`
	ApproverID   *string `json:"approver_id"`
	SLAHours     int     `json:"sla_hours"`
}
