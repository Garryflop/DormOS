package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Garryflop/DormManage/document-service/internal/domain"
	"github.com/Garryflop/DormManage/document-service/internal/service"

	documentv1 "github.com/Garryflop/DormOS-gen-go/document/v1"
)

type Server struct {
	documentv1.UnimplementedDocumentServiceServer
	svc *service.DocumentService
}

func NewServer(svc *service.DocumentService) *Server {
	return &Server{svc: svc}
}

func (s *Server) SubmitRequest(ctx context.Context, req *documentv1.SubmitDocumentRequest) (*documentv1.SubmitDocumentResponse, error) {
	studentID, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid student_id")
	}
	dormID, err := uuid.Parse(req.DormId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dorm_id")
	}

	metadata := make(map[string]string)
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	if req.Purpose != "" {
		metadata["purpose"] = req.Purpose
	}

	result, err := s.svc.SubmitRequest(ctx, studentID, dormID, req.DocumentType, metadata)
	if err != nil {
		return nil, mapError(err)
	}

	return &documentv1.SubmitDocumentResponse{
		RequestId:         result.Request.ID.String(),
		Status:            string(result.Request.Status),
		TotalSteps:        int32(result.TotalSteps),
		SubmittedAt:       result.Request.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
		FirstApproverRole: result.FirstApproverRole,
	}, nil
}

func (s *Server) ApproveStep(ctx context.Context, req *documentv1.ApproveStepRequest) (*documentv1.ApproveStepResponse, error) {
	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request_id")
	}
	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid approver_id")
	}

	result, err := s.svc.ApproveStep(ctx, requestID, approverID, req.Comment)
	if err != nil {
		return nil, mapError(err)
	}

	return &documentv1.ApproveStepResponse{
		RequestId:        result.Request.ID.String(),
		StepApproved:     int32(result.StepApproved),
		IsFinal:          result.IsFinal,
		Status:           string(result.Request.Status),
		NextApproverRole: result.NextApproverRole,
		NextApproverId:   result.NextApproverID,
		DecidedAt:        result.Request.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *Server) RejectStep(ctx context.Context, req *documentv1.RejectStepRequest) (*documentv1.RejectStepResponse, error) {
	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request_id")
	}
	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid approver_id")
	}

	result, err := s.svc.RejectStep(ctx, requestID, approverID, req.RejectionReason)
	if err != nil {
		return nil, mapError(err)
	}

	return &documentv1.RejectStepResponse{
		RequestId:  result.Request.ID.String(),
		StepNumber: int32(result.StepNumber),
		Status:     string(result.Request.Status),
		DecidedAt:  result.Request.CompletedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *Server) GetRequestStatus(ctx context.Context, req *documentv1.GetRequestStatusRequest) (*documentv1.RequestStatusResponse, error) {
	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request_id")
	}
	dormID, err := uuid.Parse(req.DormId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dorm_id")
	}

	result, err := s.svc.GetRequestStatus(ctx, requestID, dormID)
	if err != nil {
		return nil, mapError(err)
	}

	protoSteps := make([]*documentv1.ApprovalStepStatus, len(result.Steps))
	for i, step := range result.Steps {
		ps := &documentv1.ApprovalStepStatus{
			StepNumber:      int32(step.StepNumber),
			ApproverRole:    step.ApproverRole,
			Status:          string(step.Status),
			DecisionComment: step.DecisionComment,
			CreatedAt:       step.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if step.ApproverID != nil {
			ps.ApproverId = step.ApproverID.String()
		}
		if step.DecidedBy != nil {
			ps.DecidedBy = step.DecidedBy.String()
		}
		if step.DecidedAt != nil {
			ps.DecidedAt = step.DecidedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		protoSteps[i] = ps
	}

	completedAt := ""
	if result.Request.CompletedAt != nil {
		completedAt = result.Request.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &documentv1.RequestStatusResponse{
		RequestId:    result.Request.ID.String(),
		StudentId:    result.Request.StudentID.String(),
		DocumentType: result.Request.DocumentType,
		Status:       string(result.Request.Status),
		CurrentStep:  int32(result.Request.CurrentStep),
		TotalSteps:   int32(result.TotalSteps),
		SubmittedAt:  result.Request.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt:  completedAt,
		Steps:        protoSteps,
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrWorkflowNotFound),
		errors.Is(err, domain.ErrRequestNotFound),
		errors.Is(err, domain.ErrStepNotFound):
		return status.Errorf(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrRequestTerminated),
		errors.Is(err, domain.ErrAlreadyDecided):
		return status.Errorf(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrUnauthorizedStep):
		return status.Errorf(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrDocTypeRequired),
		errors.Is(err, domain.ErrStudentIDRequired):
		return status.Errorf(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
