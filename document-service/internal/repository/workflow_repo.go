package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Garryflop/DormManage/document-service/internal/domain"
)

type WorkflowRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowRepository(db *pgxpool.Pool) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) GetByDocumentType(ctx context.Context, docType string) (*domain.WorkflowDefinition, error) {
	query := `
		SELECT id, document_type, name, description, steps_json, is_active, created_at, updated_at
		FROM workflow_definitions
		WHERE document_type = $1 AND is_active = true`

	row := r.db.QueryRow(ctx, query, docType)

	var wf domain.WorkflowDefinition
	var stepsJSON []byte

	err := row.Scan(&wf.ID, &wf.DocumentType, &wf.Name, &wf.Description, &stepsJSON, &wf.IsActive, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("get workflow: %w", err)
	}

	if err := json.Unmarshal(stepsJSON, &wf.Steps); err != nil {
		return nil, fmt.Errorf("unmarshal steps: %w", err)
	}
	return &wf, nil
}

type RequestRepository struct {
	db *pgxpool.Pool
}

func NewRequestRepository(db *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{db: db}
}

func (r *RequestRepository) Create(ctx context.Context, req *domain.DocumentRequest) (*domain.DocumentRequest, error) {
	metaJSON, err := domain.MarshalMetadata(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO document_requests
			(id, student_id, workflow_id, document_type, status, current_step, metadata, submitted_at, dorm_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, student_id, workflow_id, document_type, status, current_step, metadata, submitted_at, completed_at, dorm_id`

	row := r.db.QueryRow(ctx, query,
		req.ID, req.StudentID, req.WorkflowID, req.DocumentType,
		string(req.Status), req.CurrentStep, metaJSON, req.SubmittedAt, req.DormID,
	)
	return scanRequest(row)
}

func (r *RequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DocumentRequest, error) {
	query := `
		SELECT id, student_id, workflow_id, document_type, status, current_step,
		       metadata, submitted_at, completed_at, dorm_id
		FROM document_requests WHERE id = $1`
	return scanRequest(r.db.QueryRow(ctx, query, id))
}

func (r *RequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RequestStatus, currentStep int, completedAt *time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE document_requests SET status = $1, current_step = $2, completed_at = $3 WHERE id = $4`,
		string(status), currentStep, completedAt, id,
	)
	return err
}

func scanRequest(row pgx.Row) (*domain.DocumentRequest, error) {
	var req domain.DocumentRequest
	var statusStr string
	var metaJSON []byte

	err := row.Scan(
		&req.ID, &req.StudentID, &req.WorkflowID, &req.DocumentType,
		&statusStr, &req.CurrentStep, &metaJSON, &req.SubmittedAt, &req.CompletedAt, &req.DormID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRequestNotFound
		}
		return nil, fmt.Errorf("scan request: %w", err)
	}

	req.Status = domain.RequestStatus(statusStr)
	req.Metadata, _ = domain.UnmarshalMetadata(metaJSON)
	return &req, nil
}

type StepRepository struct {
	db *pgxpool.Pool
}

func NewStepRepository(db *pgxpool.Pool) *StepRepository {
	return &StepRepository{db: db}
}

func (r *StepRepository) CreateBatch(ctx context.Context, steps []*domain.ApprovalStep) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, step := range steps {
		_, err := tx.Exec(ctx, `
			INSERT INTO approval_steps
				(id, request_id, step_number, approver_role, approver_id, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			step.ID, step.RequestID, step.StepNumber,
			step.ApproverRole, step.ApproverID, string(step.Status), step.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert step %d: %w", step.StepNumber, err)
		}
	}
	return tx.Commit(ctx)
}

func (r *StepRepository) GetByRequestID(ctx context.Context, requestID uuid.UUID) ([]*domain.ApprovalStep, error) {
	query := `
		SELECT id, request_id, step_number, approver_role, approver_id,
		       status, decided_by, decision_comment, decided_at, created_at
		FROM approval_steps
		WHERE request_id = $1
		ORDER BY step_number ASC`

	rows, err := r.db.Query(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*domain.ApprovalStep
	for rows.Next() {
		step, err := scanStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (r *StepRepository) UpdateStep(ctx context.Context, stepID uuid.UUID, status domain.StepStatus, decidedBy uuid.UUID, comment string, decidedAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE approval_steps SET status = $1, decided_by = $2, decision_comment = $3, decided_at = $4 WHERE id = $5`,
		string(status), decidedBy, comment, decidedAt, stepID,
	)
	return err
}

func scanStep(row pgx.Row) (*domain.ApprovalStep, error) {
	var s domain.ApprovalStep
	var statusStr string

	err := row.Scan(
		&s.ID, &s.RequestID, &s.StepNumber, &s.ApproverRole, &s.ApproverID,
		&statusStr, &s.DecidedBy, &s.DecisionComment, &s.DecidedAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan step: %w", err)
	}
	s.Status = domain.StepStatus(statusStr)
	return &s, nil
}
