package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Garryflop/DormManage/document-service/internal/domain"
	"github.com/Garryflop/DormManage/document-service/internal/nats"
	"github.com/Garryflop/DormManage/document-service/internal/repository"
)

type DocumentService struct {
	workflowRepo *repository.WorkflowRepository
	requestRepo  *repository.RequestRepository
	stepRepo     *repository.StepRepository
	publisher    *nats.Publisher
}

func NewDocumentService(
	workflowRepo *repository.WorkflowRepository,
	requestRepo *repository.RequestRepository,
	stepRepo *repository.StepRepository,
	publisher *nats.Publisher,
) *DocumentService {
	return &DocumentService{
		workflowRepo: workflowRepo,
		requestRepo:  requestRepo,
		stepRepo:     stepRepo,
		publisher:    publisher,
	}
}

type SubmitResult struct {
	Request           *domain.DocumentRequest
	TotalSteps        int
	FirstApproverRole string
}

func (s *DocumentService) SubmitRequest(ctx context.Context, studentID, dormID uuid.UUID, docType string, metadata map[string]string) (*SubmitResult, error) {
	if docType == "" {
		return nil, domain.ErrDocTypeRequired
	}

	workflow, err := s.workflowRepo.GetByDocumentType(ctx, docType)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	req := &domain.DocumentRequest{
		ID:           uuid.New(),
		StudentID:    studentID,
		WorkflowID:   workflow.ID,
		DocumentType: docType,
		Status:       domain.StatusInProgress,
		CurrentStep:  1,
		Metadata:     metadata,
		SubmittedAt:  now,
		DormID:       dormID,
	}

	created, err := s.requestRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	steps := make([]*domain.ApprovalStep, len(workflow.Steps))
	for i, def := range workflow.Steps {
		step := &domain.ApprovalStep{
			ID:           uuid.New(),
			RequestID:    created.ID,
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

	if err := s.stepRepo.CreateBatch(ctx, steps); err != nil {
		return nil, fmt.Errorf("create approval steps: %w", err)
	}

	firstApproverRole := workflow.Steps[0].ApproverRole
	var firstApproverID string
	if workflow.Steps[0].ApproverID != nil {
		firstApproverID = *workflow.Steps[0].ApproverID
	}

	_ = s.publisher.PublishDocumentSubmitted(ctx, nats.DocumentSubmittedEvent{
		RequestID:         created.ID.String(),
		StudentID:         studentID.String(),
		DocumentType:      docType,
		FirstApproverID:   firstApproverID,
		FirstApproverRole: firstApproverRole,
	})

	return &SubmitResult{
		Request:           created,
		TotalSteps:        len(workflow.Steps),
		FirstApproverRole: firstApproverRole,
	}, nil
}

type ApproveResult struct {
	Request          *domain.DocumentRequest
	StepApproved     int
	IsFinal          bool
	NextApproverRole string
	NextApproverID   string
}

func (s *DocumentService) ApproveStep(ctx context.Context, requestID, approverID uuid.UUID, comment string) (*ApproveResult, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.Status != domain.StatusInProgress {
		return nil, domain.ErrRequestTerminated
	}

	steps, err := s.stepRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	currentStep := findStep(steps, req.CurrentStep)
	if currentStep == nil {
		return nil, domain.ErrStepNotFound
	}

	if currentStep.Status != domain.StepPending {
		return nil, domain.ErrAlreadyDecided
	}

	if currentStep.ApproverID != nil && *currentStep.ApproverID != approverID {
		return nil, domain.ErrUnauthorizedStep
	}

	now := time.Now()
	if err := s.stepRepo.UpdateStep(ctx, currentStep.ID, domain.StepApproved, approverID, comment, now); err != nil {
		return nil, fmt.Errorf("update step: %w", err)
	}

	nextStep := findStep(steps, req.CurrentStep+1)
	isFinal := nextStep == nil

	var newStatus domain.RequestStatus
	var nextCurrentStep int
	var completedAt *time.Time

	if isFinal {
		newStatus = domain.StatusApproved
		nextCurrentStep = req.CurrentStep
		completedAt = &now
	} else {
		newStatus = domain.StatusInProgress
		nextCurrentStep = req.CurrentStep + 1
	}

	if err := s.requestRepo.UpdateStatus(ctx, requestID, newStatus, nextCurrentStep, completedAt); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}
	req.Status = newStatus
	req.CurrentStep = nextCurrentStep
	req.CompletedAt = completedAt

	var nextApproverRole, nextApproverID string
	if !isFinal {
		nextApproverRole = nextStep.ApproverRole
		if nextStep.ApproverID != nil {
			nextApproverID = nextStep.ApproverID.String()
		}
	}

	_ = s.publisher.PublishDocumentStepApproved(ctx, nats.DocumentStepApprovedEvent{
		RequestID:        requestID.String(),
		StudentID:        req.StudentID.String(),
		StepNumber:       req.CurrentStep,
		ApproverID:       approverID.String(),
		NextApproverRole: nextApproverRole,
		NextApproverID:   nextApproverID,
		IsFinal:          isFinal,
	})

	return &ApproveResult{
		Request:          req,
		StepApproved:     currentStep.StepNumber,
		IsFinal:          isFinal,
		NextApproverRole: nextApproverRole,
		NextApproverID:   nextApproverID,
	}, nil
}

type RejectResult struct {
	Request    *domain.DocumentRequest
	StepNumber int
}

func (s *DocumentService) RejectStep(ctx context.Context, requestID, approverID uuid.UUID, reason string) (*RejectResult, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.Status != domain.StatusInProgress {
		return nil, domain.ErrRequestTerminated
	}

	steps, err := s.stepRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	currentStep := findStep(steps, req.CurrentStep)
	if currentStep == nil {
		return nil, domain.ErrStepNotFound
	}

	if currentStep.Status != domain.StepPending {
		return nil, domain.ErrAlreadyDecided
	}

	now := time.Now()
	if err := s.stepRepo.UpdateStep(ctx, currentStep.ID, domain.StepRejected, approverID, reason, now); err != nil {
		return nil, fmt.Errorf("update step: %w", err)
	}

	if err := s.requestRepo.UpdateStatus(ctx, requestID, domain.StatusRejected, req.CurrentStep, &now); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}
	req.Status = domain.StatusRejected
	req.CompletedAt = &now

	_ = s.publisher.PublishDocumentStepRejected(ctx, nats.DocumentStepRejectedEvent{
		RequestID:       requestID.String(),
		StudentID:       req.StudentID.String(),
		StepNumber:      currentStep.StepNumber,
		ApproverID:      approverID.String(),
		RejectionReason: reason,
	})

	return &RejectResult{Request: req, StepNumber: currentStep.StepNumber}, nil
}

type RequestStatusResult struct {
	Request    *domain.DocumentRequest
	Steps      []*domain.ApprovalStep
	TotalSteps int
}

func (s *DocumentService) GetRequestStatus(ctx context.Context, requestID, dormID uuid.UUID) (*RequestStatusResult, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.DormID != dormID {
		return nil, domain.ErrRequestNotFound
	}

	steps, err := s.stepRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	return &RequestStatusResult{Request: req, Steps: steps, TotalSteps: len(steps)}, nil
}

func findStep(steps []*domain.ApprovalStep, stepNumber int) *domain.ApprovalStep {
	for _, s := range steps {
		if s.StepNumber == stepNumber {
			return s
		}
	}
	return nil
}
