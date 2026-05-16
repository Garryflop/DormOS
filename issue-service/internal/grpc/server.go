package grpc

import (
	"context"
	"fmt"

	issuev1 "github.com/Garryflop/DormOS-gen-go/issue/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Garryflop/DormManage/issue-service/internal/domain"
	"github.com/Garryflop/DormManage/issue-service/internal/service"
)

type IssueServer struct {
	issuev1.UnimplementedIssueServiceServer
	svc *service.IssueService
}

func NewIssueServer(svc *service.IssueService) *IssueServer {
	return &IssueServer{svc: svc}
}

// 1. CreateIssue
func (s *IssueServer) CreateIssue(ctx context.Context, req *issuev1.CreateIssueRequest) (*issuev1.CreateIssueResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	catID, err := uuid.Parse(req.CategoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid category_id: %v", err)
	}

	var photoURL *string
	if len(req.PhotoUrls) > 0 {
		photoURL = &req.PhotoUrls[0]
	}

	issue, err := s.svc.CreateIssue(ctx, domain.CreateIssueInput{
		UserID:      userID,
		RoomNumber:  req.RoomNumber,
		CategoryID:  catID,
		Title:       req.Title,
		Description: req.Description,
		PhotoURL:    photoURL,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create issue: %v", err)
	}

	return &issuev1.CreateIssueResponse{IssueId: issue.ID.String()}, nil
}

// 2. GetIssue
func (s *IssueServer) GetIssue(ctx context.Context, req *issuev1.GetIssueRequest) (*issuev1.GetIssueResponse, error) {
	id, err := uuid.Parse(req.IssueId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue_id")
	}

	issue, err := s.svc.GetIssue(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "issue not found: %v", err)
	}

	resp := &issuev1.GetIssueResponse{
		IssueId:     issue.ID.String(),
		UserId:      issue.UserID.String(),
		RoomNumber:  issue.RoomNumber,
		Category:    issue.CategoryID.String(),
		Title:       issue.Title,
		Description: issue.Description,
		Status:      string(issue.Status),
		CreatedAt:   issue.CreatedAt.Unix(),
		UpdatedAt:   issue.UpdatedAt.Unix(),
	}
	if issue.PhotoURL != nil {
		resp.PhotoUrls = []string{*issue.PhotoURL}
	}
	return resp, nil
}

// 3. ListMyIssues
func (s *IssueServer) ListMyIssues(ctx context.Context, req *issuev1.ListMyIssuesRequest) (*issuev1.ListMyIssuesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}

	issues, err := s.svc.ListMyIssues(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var items []*issuev1.GetIssueResponse
	for _, i := range issues {
		item := &issuev1.GetIssueResponse{
			IssueId:     i.ID.String(),
			UserId:      i.UserID.String(),
			RoomNumber:  i.RoomNumber,
			Title:       i.Title,
			Description: i.Description,
			Status:      string(i.Status),
			CreatedAt:   i.CreatedAt.Unix(),
		}
		items = append(items, item)
	}

	return &issuev1.ListMyIssuesResponse{Issues: items}, nil
}

// 4. ListAllIssues
func (s *IssueServer) ListAllIssues(ctx context.Context, _ *issuev1.ListAllIssuesRequest) (*issuev1.ListAllIssuesResponse, error) {
	issues, err := s.svc.ListAllIssues(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var items []*issuev1.GetIssueResponse
	for _, i := range issues {
		items = append(items, &issuev1.GetIssueResponse{
			IssueId:    i.ID.String(),
			UserId:     i.UserID.String(),
			RoomNumber: i.RoomNumber,
			Title:      i.Title,
			Status:     string(i.Status),
			CreatedAt:  i.CreatedAt.Unix(),
		})
	}

	return &issuev1.ListAllIssuesResponse{Issues: items}, nil
}

// 5. UpdateIssueStatus
func (s *IssueServer) UpdateIssueStatus(ctx context.Context, req *issuev1.UpdateIssueStatusRequest) (*issuev1.UpdateIssueStatusResponse, error) {
	issueID, err := uuid.Parse(req.IssueId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue_id")
	}

	_, err = s.svc.UpdateIssueStatus(ctx, domain.UpdateIssueStatusInput{
		IssueID: issueID,
		Status:  domain.Status(req.NewStatus),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &issuev1.UpdateIssueStatusResponse{Success: true}, nil
}

// 6. DeleteIssue
func (s *IssueServer) DeleteIssue(ctx context.Context, req *issuev1.DeleteIssueRequest) (*issuev1.DeleteIssueResponse, error) {
	id, err := uuid.Parse(req.IssueId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue_id")
	}

	if err := s.svc.DeleteIssue(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &issuev1.DeleteIssueResponse{Success: true}, nil
}

// 7. AddComment
func (s *IssueServer) AddComment(ctx context.Context, req *issuev1.AddCommentRequest) (*issuev1.AddCommentResponse, error) {
	issueID, err := uuid.Parse(req.IssueId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue_id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}

	comment, err := s.svc.AddComment(ctx, domain.AddCommentInput{
		IssueID: issueID,
		UserID:  userID,
		Text:    req.Content,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &issuev1.AddCommentResponse{
		CommentId: comment.ID.String(),
	}, nil
}

// 8. ListComments
func (s *IssueServer) ListComments(ctx context.Context, req *issuev1.ListCommentsRequest) (*issuev1.ListCommentsResponse, error) {
	issueID, err := uuid.Parse(req.IssueId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue_id")
	}

	comments, err := s.svc.ListComments(ctx, issueID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var items []*issuev1.CommentItem
	for _, c := range comments {
		items = append(items, &issuev1.CommentItem{
			CommentId: c.ID.String(),
			UserId:    c.UserID.String(),
			Content:   c.Text,
			CreatedAt: c.CreatedAt.Unix(),
		})
	}

	return &issuev1.ListCommentsResponse{Comments: items}, nil
}

// 9. AssignWorker
func (s *IssueServer) AssignWorker(ctx context.Context, req *issuev1.AssignWorkerRequest) (*issuev1.AssignWorkerResponse, error) {
	issueID, err := uuid.Parse(req.IssueId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue_id")
	}

	// Find worker by name
	workers, err := s.svc.ListWorkers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var workerID uuid.UUID
	for _, w := range workers {
		if w.Name == req.WorkerName {
			workerID = w.ID
			break
		}
	}
	if workerID == uuid.Nil {
		return nil, status.Errorf(codes.NotFound, "worker not found: %s", req.WorkerName)
	}

	_, err = s.svc.AssignWorker(ctx, issueID, workerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &issuev1.AssignWorkerResponse{Success: true}, nil
}

// 10. ListWorkers
func (s *IssueServer) ListWorkers(ctx context.Context, _ *issuev1.ListWorkersRequest) (*issuev1.ListWorkersResponse, error) {
	workers, err := s.svc.ListWorkers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var items []*issuev1.WorkerItem
	for _, w := range workers {
		items = append(items, &issuev1.WorkerItem{
			WorkerId:       w.ID.String(),
			Name:           w.Name,
			Specialization: w.Specialty,
		})
	}

	return &issuev1.ListWorkersResponse{Workers: items}, nil
}

// 11. CreateCategory
func (s *IssueServer) CreateCategory(ctx context.Context, req *issuev1.CreateCategoryRequest) (*issuev1.CreateCategoryResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	cat, err := s.svc.CreateCategory(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &issuev1.CreateCategoryResponse{CategoryId: cat.ID.String()}, nil
}

// 12. ListCategories
func (s *IssueServer) ListCategories(ctx context.Context, _ *issuev1.ListCategoriesRequest) (*issuev1.ListCategoriesResponse, error) {
	cats, err := s.svc.ListCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var items []*issuev1.CategoryItem
	for _, c := range cats {
		items = append(items, &issuev1.CategoryItem{
			CategoryId: c.ID.String(),
			Name:       c.Name,
		})
	}

	return &issuev1.ListCategoriesResponse{Categories: items}, nil
}

// ensure interface compliance
var _ issuev1.IssueServiceServer = (*IssueServer)(nil)

// suppress unused import
var _ = fmt.Sprintf
