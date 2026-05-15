package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	StatusInProgress RequestStatus = "in_progress"
	StatusApproved   RequestStatus = "approved"
	StatusRejected   RequestStatus = "rejected"
	StatusCancelled  RequestStatus = "cancelled"
)

type StepStatus string

const (
	StepPending  StepStatus = "pending"
	StepApproved StepStatus = "approved"
	StepRejected StepStatus = "rejected"
	StepSkipped  StepStatus = "skipped"
)

type WorkflowStepDef struct {
	StepNumber   int     `json:"step_number"`
	ApproverRole string  `json:"approver_role"`
	ApproverID   *string `json:"approver_id"`
	SLAHours     int     `json:"sla_hours"`
}

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

type DocumentRequest struct {
	ID           uuid.UUID
	StudentID    uuid.UUID
	WorkflowID   uuid.UUID
	DocumentType string
	Status       RequestStatus
	CurrentStep  int
	Metadata     map[string]string
	SubmittedAt  time.Time
	CompletedAt  *time.Time
	DormID       uuid.UUID
}

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

var (
	ErrWorkflowNotFound  = errors.New("workflow definition not found for this document type")
	ErrRequestNotFound   = errors.New("document request not found")
	ErrStepNotFound      = errors.New("approval step not found")
	ErrAlreadyDecided    = errors.New("this step has already been decided")
	ErrRequestTerminated = errors.New("request has already reached a terminal state")
	ErrUnauthorizedStep  = errors.New("you are not authorized to approve this step")
	ErrDocTypeRequired   = errors.New("document_type is required")
	ErrStudentIDRequired = errors.New("student_id is required")
)

func MarshalMetadata(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func UnmarshalMetadata(data []byte) (map[string]string, error) {
	if len(data) == 0 {
		return map[string]string{}, nil
	}
	var m map[string]string
	return m, json.Unmarshal(data, &m)
}
