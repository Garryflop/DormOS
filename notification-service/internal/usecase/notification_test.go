package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/dormos/notification-service/internal/domain"
	"github.com/dormos/notification-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock Notification Repository ─────────────────────────────────────────────

type mockNotificationRepo struct{ mock.Mock }

func (m *mockNotificationRepo) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	args := m.Called(ctx, n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
}

func (m *mockNotificationRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.Notification, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Notification), args.Error(1)
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockNotificationRepo) GetUserEmail(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

// ─── Test Notification Usecase ────────────────────────────────────────────────

func TestNotificationUseCase_SendNotification_Success(t *testing.T) {
	repo := &mockNotificationRepo{}
	cfg := usecase.SMTPConfig{
		Host:     "localhost",
		Port:     "25",
		Username: "user",
		Password: "password",
		From:     "noreply@dormos.com",
	}
	uc := usecase.NewNotificationUseCase(repo, cfg)

	expectedNotif := &domain.Notification{
		UserID:  "user-1",
		Title:   "Test Alert",
		Message: "This is a test notification message",
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Run(func(args mock.Arguments) {
			n := args.Get(1).(*domain.Notification)
			assert.Equal(t, "user-1", n.UserID)
			assert.Equal(t, "Test Alert", n.Title)
			assert.Equal(t, "This is a test notification message", n.Message)
		}).
		Return(expectedNotif, nil)

	err := uc.SendNotification(context.Background(), "user-1", "test@user.com", "Test Alert", "This is a test notification message")
	assert.NoError(t, err)

	// Wait briefly to allow goroutine for sendEmail to execute (even if it fails to dial, it doesn't block the test)
	time.Sleep(50 * time.Millisecond)

	repo.AssertExpectations(t)
}

func TestNotificationUseCase_SendNotification_FetchEmailSuccess(t *testing.T) {
	repo := &mockNotificationRepo{}
	cfg := usecase.SMTPConfig{
		Host:     "localhost",
		Port:     "25",
		Username: "user",
		Password: "password",
		From:     "noreply@dormos.com",
	}
	uc := usecase.NewNotificationUseCase(repo, cfg)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(&domain.Notification{}, nil)
	repo.On("GetUserEmail", mock.Anything, "user-1").Return("fetched@user.com", nil)

	err := uc.SendNotification(context.Background(), "user-1", "", "Test Alert", "Message text")
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	repo.AssertExpectations(t)
}

func TestNotificationUseCase_ListMyNotifications_Success(t *testing.T) {
	repo := &mockNotificationRepo{}
	uc := usecase.NewNotificationUseCase(repo, usecase.SMTPConfig{})

	expected := []*domain.Notification{
		{ID: "n-1", UserID: "user-1", Title: "Title 1", Message: "Msg 1"},
		{ID: "n-2", UserID: "user-1", Title: "Title 2", Message: "Msg 2"},
	}

	repo.On("ListByUserID", mock.Anything, "user-1").Return(expected, nil)

	res, err := uc.ListMyNotifications(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Title 1", res[0].Title)
}

func TestNotificationUseCase_ReadOperations_Success(t *testing.T) {
	repo := &mockNotificationRepo{}
	uc := usecase.NewNotificationUseCase(repo, usecase.SMTPConfig{})

	repo.On("MarkAsRead", mock.Anything, "n-1").Return(nil)
	repo.On("MarkAllAsRead", mock.Anything, "user-1").Return(nil)

	err := uc.MarkAsRead(context.Background(), "n-1")
	assert.NoError(t, err)

	err = uc.MarkAllAsRead(context.Background(), "user-1")
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}
