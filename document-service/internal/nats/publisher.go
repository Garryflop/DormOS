package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}
	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Close() {
	p.conn.Drain()
}

type DocumentSubmittedEvent struct {
	RequestID         string `json:"request_id"`
	StudentID         string `json:"student_id"`
	DocumentType      string `json:"document_type"`
	FirstApproverID   string `json:"first_approver_id"`
	FirstApproverRole string `json:"first_approver_role"`
}

type DocumentStepApprovedEvent struct {
	RequestID        string `json:"request_id"`
	StudentID        string `json:"student_id"`
	StepNumber       int    `json:"step_number"`
	ApproverID       string `json:"approver_id"`
	NextApproverRole string `json:"next_approver_role"`
	NextApproverID   string `json:"next_approver_id"`
	IsFinal          bool   `json:"is_final"`
}

type DocumentStepRejectedEvent struct {
	RequestID       string `json:"request_id"`
	StudentID       string `json:"student_id"`
	StepNumber      int    `json:"step_number"`
	ApproverID      string `json:"approver_id"`
	RejectionReason string `json:"rejection_reason"`
}

func (p *Publisher) PublishDocumentSubmitted(_ context.Context, event DocumentSubmittedEvent) error {
	return p.publish("document.submitted", event)
}

func (p *Publisher) PublishDocumentStepApproved(_ context.Context, event DocumentStepApprovedEvent) error {
	return p.publish("document.step.approved", event)
}

func (p *Publisher) PublishDocumentStepRejected(_ context.Context, event DocumentStepRejectedEvent) error {
	return p.publish("document.step.rejected", event)
}

func (p *Publisher) publish(subject string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s event: %w", subject, err)
	}
	return p.conn.Publish(subject, data)
}
