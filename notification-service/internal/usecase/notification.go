package usecase

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/dormos/notification-service/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type NotificationUseCase struct {
	repo NotificationRepository
	smtp SMTPConfig
}

func NewNotificationUseCase(repo NotificationRepository, smtpConfig SMTPConfig) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, smtp: smtpConfig}
}

func (uc *NotificationUseCase) SendNotification(ctx context.Context, userID, email, title, message string) error {
	notif := &domain.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
	}
	if _, err := uc.repo.Create(ctx, notif); err != nil {
		return fmt.Errorf("NotificationUseCase.SendNotification DB: %w", err)
	}

	if email == "" {
		if fetchedEmail, err := uc.repo.GetUserEmail(ctx, userID); err == nil {
			email = fetchedEmail
		} else {
			log.Printf("[SMTP] failed to lookup email for userID %s: %v", userID, err)
		}
	}

	if email != "" {
		go func() {
			if err := uc.sendEmail(email, title, message); err != nil {
				log.Printf("[SMTP] failed to send email to %s: %v", email, err)
			}
		}()
	}
	return nil
}

func (uc *NotificationUseCase) sendEmail(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", uc.smtp.Host, uc.smtp.Port)
	auth := smtp.PlainAuth("", uc.smtp.Username, uc.smtp.Password, uc.smtp.Host)

	msg := fmt.Sprintf(
		"From: DormOS <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		uc.smtp.From, to, subject, body,
	)

	if uc.smtp.Port == "465" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         uc.smtp.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, uc.smtp.Host)
		if err != nil {
			return fmt.Errorf("SMTP NewClient: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP Auth: %w", err)
		}
		if err = client.Mail(uc.smtp.From); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, uc.smtp.From, []string{to}, []byte(msg))
}

func (uc *NotificationUseCase) ListMyNotifications(ctx context.Context, userID string) ([]*domain.Notification, error) {
	notifs, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("NotificationUseCase.ListMyNotifications: %w", err)
	}
	return notifs, nil
}

func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, id string) error {
	if err := uc.repo.MarkAsRead(ctx, id); err != nil {
		return fmt.Errorf("NotificationUseCase.MarkAsRead: %w", err)
	}
	return nil
}

func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, userID string) error {
	if err := uc.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("NotificationUseCase.MarkAllAsRead: %w", err)
	}
	return nil
}
