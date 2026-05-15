package nats

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectIssueCreated       = "issue.created"
	SubjectIssueStatusChanged = "issue.status_changed"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	log.Println("[NATS] Connected to", url)
	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Close() {
	p.conn.Drain()
}

type IssueCreatedEvent struct {
	IssueID    uuid.UUID `json:"issue_id"`
	UserID     uuid.UUID `json:"user_id"`
	RoomNumber string    `json:"room_number"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
}

type IssueStatusChangedEvent struct {
	IssueID   uuid.UUID `json:"issue_id"`
	UserID    uuid.UUID `json:"user_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
}

func (p *Publisher) PublishIssueCreated(_ context.Context, ev IssueCreatedEvent) {
	p.publish(SubjectIssueCreated, ev)
}

func (p *Publisher) PublishIssueStatusChanged(_ context.Context, ev IssueStatusChangedEvent) {
	p.publish(SubjectIssueStatusChanged, ev)
}

func (p *Publisher) publish(subject string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[NATS] marshal error %s: %v", subject, err)
		return
	}
	if err := p.conn.Publish(subject, data); err != nil {
		log.Printf("[NATS] publish error %s: %v", subject, err)
	}
}
