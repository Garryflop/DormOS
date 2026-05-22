package nats

import (
	"context"
	"encoding/json"
	"log"

	"github.com/dormos/notification-service/internal/domain"
	"github.com/dormos/notification-service/internal/usecase"

	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	nc             *nats.Conn
	notificationUC *usecase.NotificationUseCase
}

func NewSubscriber(nc *nats.Conn, notifUC *usecase.NotificationUseCase) *Subscriber {
	return &Subscriber{nc: nc, notificationUC: notifUC}
}

func (s *Subscriber) Subscribe() error {
	subjects := []struct {
		subject string
		handler nats.MsgHandler
	}{
		{"issue.created", s.handleIssueCreated},
		{"issue.status_changed", s.handleIssueStatusChanged},
		{"event.created", s.handleEventCreated},
		{"event.cancelled", s.handleEventCancelled},
		{"resident.assigned", s.handleResidentAssigned},
		{"resident.removed", s.handleResidentRemoved},
		{"password.reset", s.handlePasswordReset},
		{"role.updated", s.handleRoleUpdated},
	}

	for _, sub := range subjects {
		if _, err := s.nc.Subscribe(sub.subject, sub.handler); err != nil {
			return err
		}
		log.Printf("[NATS] subscribed to subject: %s", sub.subject)
	}
	return nil
}

func (s *Subscriber) handleIssueCreated(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][issue.created] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()
	err := s.notificationUC.SendNotification(ctx,
		p.UserID,
		p.Email,
		"New Issue Created",
		p.Message,
	)
	if err != nil {
		log.Printf("[NATS][issue.created] SendNotification error: %v", err)
	}
}

func (s *Subscriber) handleIssueStatusChanged(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][issue.status_changed] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()
	err := s.notificationUC.SendNotification(ctx,
		p.UserID,
		p.Email,
		"Issue Status Changed",
		p.Message,
	)
	if err != nil {
		log.Printf("[NATS][issue.status_changed] SendNotification error: %v", err)
	}
}

func (s *Subscriber) handleEventCreated(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][event.created] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()

	for i, uid := range p.UserIDs {
		email := ""
		if i < len(p.Emails) {
			email = p.Emails[i]
		}
		if err := s.notificationUC.SendNotification(ctx, uid, email, p.Title, p.Message); err != nil {
			log.Printf("[NATS][event.created] SendNotification uid=%s error: %v", uid, err)
		}
	}
}

func (s *Subscriber) handleEventCancelled(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][event.cancelled] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()

	for i, uid := range p.UserIDs {
		email := ""
		if i < len(p.Emails) {
			email = p.Emails[i]
		}
		if err := s.notificationUC.SendNotification(ctx, uid, email, p.Title, p.Message); err != nil {
			log.Printf("[NATS][event.cancelled] SendNotification uid=%s error: %v", uid, err)
		}
	}
}

func (s *Subscriber) handleResidentAssigned(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][resident.assigned] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()
	
	title := p.Title
	if title == "" {
		title = "Room Assigned"
	}
	msgText := p.Message
	if msgText == "" {
		msgText = "You have been successfully assigned to a room."
	}

	err := s.notificationUC.SendNotification(ctx,
		p.UserID,
		p.Email,
		title,
		msgText,
	)
	if err != nil {
		log.Printf("[NATS][resident.assigned] SendNotification error: %v", err)
	}
}

func (s *Subscriber) handleResidentRemoved(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][resident.removed] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()

	title := p.Title
	if title == "" {
		title = "Evicted from Room"
	}
	msgText := p.Message
	if msgText == "" {
		msgText = "You have been successfully evicted from your room by the administrator."
	}

	err := s.notificationUC.SendNotification(ctx,
		p.UserID,
		p.Email,
		title,
		msgText,
	)
	if err != nil {
		log.Printf("[NATS][resident.removed] SendNotification error: %v", err)
	}
}

func (s *Subscriber) handlePasswordReset(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][password.reset] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()
	err := s.notificationUC.SendNotification(ctx,
		p.UserID,
		p.Email,
		"DormOS Password Reset",
		p.Message,
	)
	if err != nil {
		log.Printf("[NATS][password.reset] SendNotification error: %v", err)
	}
}

func (s *Subscriber) handleRoleUpdated(msg *nats.Msg) {
	var p domain.NATSEventPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		log.Printf("[NATS][role.updated] unmarshal error: %v", err)
		return
	}
	ctx := context.Background()

	title := p.Title
	if title == "" {
		title = "System Role Updated"
	}
	msgText := p.Message
	if msgText == "" {
		msgText = "The administrator has updated your system role."
	}

	err := s.notificationUC.SendNotification(ctx,
		p.UserID,
		p.Email,
		title,
		msgText,
	)
	if err != nil {
		log.Printf("[NATS][role.updated] SendNotification error: %v", err)
	}
}
