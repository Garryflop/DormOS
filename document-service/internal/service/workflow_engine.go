package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Garryflop/DormManage/document-service/internal/domain"
	"github.com/Garryflop/DormManage/document-service/internal/repository"
)

type WorkflowEngine struct {
	workflowRepo *repository.WorkflowRepository
	requestRepo  *repository.RequestRepository
	stepRepo     *repository.StepRepository
}

func NewWorkflowEngine(
	workflowRepo *repository.WorkflowRepository,
	requestRepo *repository.RequestRepository,
	stepRepo *repository.StepRepository,
) *WorkflowEngine {
	return &WorkflowEngine{
		workflowRepo: workflowRepo,
		requestRepo:  requestRepo,
		stepRepo:     stepRepo,
	}
}

func (e *WorkflowEngine) InitializeWorkflow(ctx context.Context, requestID uuid.UUID, docType string) ([]*domain.ApprovalStep, error) {
	workflow, err := e.workflowRepo.GetByDocumentType(ctx, docType)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	steps := make([]*domain.ApprovalStep, len(workflow.Steps))
	for i, def := range workflow.Steps {
		step := &domain.ApprovalStep{
			ID:           uuid.New(),
			RequestID:    requestID,
			StepNumber:   def.StepNumber,
			ApproverRole: def.ApproverRole,
			Status:       domain.StepPending,
			CreatedAt:    now,
		}
		if def.ApproverID != nil {
			id, err := uuid.Parse(*def.ApproverID)
			if err == nil {
				step.ApproverID = &id
			}
		}
		steps[i] = step
	}

	if err := e.stepRepo.CreateBatch(ctx, steps); err != nil {
		return nil, fmt.Errorf("create steps: %w", err)
	}
	return steps, nil
}

func (e *WorkflowEngine) Advance(ctx context.Context, requestID, approverID uuid.UUID, comment string) (*AdvanceResult, error) {
	req, err := e.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.Status != domain.StatusInProgress {
		return nil, domain.ErrRequestTerminated
	}

	steps, err := e.stepRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	current := findStepByNumber(steps, req.CurrentStep)
	if current == nil {
		return nil, domain.ErrStepNotFound
	}

	if current.Status != domain.StepPending {
		return nil, domain.ErrAlreadyDecided
	}

	if current.ApproverID != nil && *current.ApproverID != approverID {
		return nil, domain.ErrUnauthorizedStep
	}

	now := time.Now()
	if err := e.stepRepo.UpdateStep(ctx, current.ID, domain.StepApproved, approverID, comment, now); err != nil {
		return nil, fmt.Errorf("update step: %w", err)
	}

	next := findStepByNumber(steps, req.CurrentStep+1)
	isFinal := next == nil

	var newStatus domain.RequestStatus
	var nextStep int
	var completedAt *time.Time

	if isFinal {
		newStatus = domain.StatusApproved
		nextStep = req.CurrentStep
		completedAt = &now
	} else {
		newStatus = domain.StatusInProgress
		nextStep = req.CurrentStep + 1
	}

	if err := e.requestRepo.UpdateStatus(ctx, requestID, newStatus, nextStep, completedAt); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	return &AdvanceResult{
		IsFinal:          isFinal,
		StepApproved:     current.StepNumber,
		NextApproverRole: nextApproverRole(next),
		NextApproverID:   nextApproverIDStr(next),
		NewStatus:        newStatus,
		NextStep:         nextStep,
		CompletedAt:      completedAt,
	}, nil
}

func (e *WorkflowEngine) Reject(ctx context.Context, requestID, approverID uuid.UUID, reason string) error {
	req, err := e.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.Status != domain.StatusInProgress {
		return domain.ErrRequestTerminated
	}

	steps, err := e.stepRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return err
	}

	current := findStepByNumber(steps, req.CurrentStep)
	if current == nil {
		return domain.ErrStepNotFound
	}

	if current.Status != domain.StepPending {
		return domain.ErrAlreadyDecided
	}

	now := time.Now()
	if err := e.stepRepo.UpdateStep(ctx, current.ID, domain.StepRejected, approverID, reason, now); err != nil {
		return fmt.Errorf("update step: %w", err)
	}

	return e.requestRepo.UpdateStatus(ctx, requestID, domain.StatusRejected, req.CurrentStep, &now)
}

type AdvanceResult struct {
	IsFinal          bool
	StepApproved     int
	NextApproverRole string
	NextApproverID   string
	NewStatus        domain.RequestStatus
	NextStep         int
	CompletedAt      *time.Time
}

func findStepByNumber(steps []*domain.ApprovalStep, n int) *domain.ApprovalStep {
	for _, s := range steps {
		if s.StepNumber == n {
			return s
		}
	}
	return nil
}

func nextApproverRole(step *domain.ApprovalStep) string {
	if step == nil {
		return ""
	}
	return step.ApproverRole
}

func nextApproverIDStr(step *domain.ApprovalStep) string {
	if step == nil || step.ApproverID == nil {
		return ""
	}
	return step.ApproverID.String()
}
