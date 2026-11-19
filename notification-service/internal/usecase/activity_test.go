package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dormos/notification-service/internal/domain"
	"github.com/dormos/notification-service/internal/usecase"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock Event Repository ───────────────────────────────────────────────────

type mockEventRepo struct{ mock.Mock }

func (m *mockEventRepo) Create(ctx context.Context, event *domain.Event) (*domain.Event, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *mockEventRepo) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *mockEventRepo) List(ctx context.Context) ([]*domain.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *mockEventRepo) Update(ctx context.Context, event *domain.Event) (*domain.Event, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *mockEventRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockEventRepo) RegisterUser(ctx context.Context, eventID, userID string) error {
	return m.Called(ctx, eventID, userID).Error(0)
}

func (m *mockEventRepo) CancelRegistration(ctx context.Context, eventID, userID string) error {
	return m.Called(ctx, eventID, userID).Error(0)
}

func (m *mockEventRepo) RecordAttendance(ctx context.Context, eventID, userID string, attended bool) error {
	return m.Called(ctx, eventID, userID, attended).Error(0)
}

func (m *mockEventRepo) GetRegisteredUserIDs(ctx context.Context, eventID string) ([]string, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ─── Mock NATS Publisher ──────────────────────────────────────────────────────

type mockNATSPublisher struct{ mock.Mock }

func (m *mockNATSPublisher) Publish(subject string, data []byte) error {
	return m.Called(subject, data).Error(0)
}

// ─── Test Activity Usecase ────────────────────────────────────────────────────

func TestActivityUseCase_CreateEvent_Success(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := &mockEventRepo{}
	pub := &mockNATSPublisher{}
	uc := usecase.NewActivityUseCase(repo, rdb, pub)

	event := &domain.Event{
		ID:          "ev-1",
		Title:       "Pizza Party",
		Description: "Free pizza!",
		Location:    "Dorm Kitchen",
		EventDate:   time.Now().Add(24 * time.Hour),
	}

	repo.On("Create", mock.Anything, event).Return(event, nil)
	pub.On("Publish", "event.created", mock.Anything).Return(nil)

	res, err := uc.CreateEvent(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, "ev-1", res.ID)

	repo.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestActivityUseCase_GetEvent_SuccessAndError(t *testing.T) {
	repo := &mockEventRepo{}
	uc := usecase.NewActivityUseCase(repo, nil, nil)

	event := &domain.Event{ID: "ev-1", Title: "Pizza Party"}

	repo.On("GetByID", mock.Anything, "ev-1").Return(event, nil)
	repo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("not found"))

	// Success
	res, err := uc.GetEvent(context.Background(), "ev-1")
	assert.NoError(t, err)
	assert.Equal(t, "Pizza Party", res.Title)

	// Error
	_, err = uc.GetEvent(context.Background(), "non-existent")
	assert.Error(t, err)
}

func TestActivityUseCase_ListEvents_CacheHitAndMiss(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := &mockEventRepo{}
	uc := usecase.NewActivityUseCase(repo, rdb, nil)

	events := []*domain.Event{
		{ID: "ev-1", Title: "Pizza Party"},
		{ID: "ev-2", Title: "Movie Night"},
	}

	// 1. Cache Miss -> Read from DB and Save to cache
	repo.On("List", mock.Anything).Return(events, nil)

	res, err := uc.ListEvents(context.Background())
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Pizza Party", res[0].Title)

	// Verify key exists in Redis
	exists := mr.Exists("events:list")
	assert.True(t, exists)

	// 2. Cache Hit -> Read from Cache (Repo List shouldn't be called again)
	res2, err := uc.ListEvents(context.Background())
	assert.NoError(t, err)
	assert.Len(t, res2, 2)
	assert.Equal(t, "Movie Night", res2[1].Title)

	repo.AssertNumberOfCalls(t, "List", 1)
}

func TestActivityUseCase_UpdateEvent_Success(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Put dummy cache
	_ = rdb.Set(context.Background(), "events:list", "cached-data", 0).Err()

	repo := &mockEventRepo{}
	uc := usecase.NewActivityUseCase(repo, rdb, nil)

	event := &domain.Event{ID: "ev-1", Title: "Updated Pizza Party"}
	repo.On("Update", mock.Anything, event).Return(event, nil)

	res, err := uc.UpdateEvent(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Pizza Party", res.Title)

	// Verify cache was invalidated
	exists := mr.Exists("events:list")
	assert.False(t, exists)
}

func TestActivityUseCase_DeleteEvent_Success(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Put dummy cache
	_ = rdb.Set(context.Background(), "events:list", "cached-data", 0).Err()

	repo := &mockEventRepo{}
	pub := &mockNATSPublisher{}
	uc := usecase.NewActivityUseCase(repo, rdb, pub)

	repo.On("GetRegisteredUserIDs", mock.Anything, "ev-1").Return([]string{"u1", "u2"}, nil)
	repo.On("Delete", mock.Anything, "ev-1").Return(nil)
	pub.On("Publish", "event.cancelled", mock.Anything).Return(nil)

	err := uc.DeleteEvent(context.Background(), "ev-1")
	assert.NoError(t, err)

	// Verify cache was invalidated
	exists := mr.Exists("events:list")
	assert.False(t, exists)

	repo.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestActivityUseCase_RegisterAndCancel_Success(t *testing.T) {
	repo := &mockEventRepo{}
	uc := usecase.NewActivityUseCase(repo, nil, nil)

	repo.On("RegisterUser", mock.Anything, "ev-1", "user-1").Return(nil)
	repo.On("CancelRegistration", mock.Anything, "ev-1", "user-1").Return(nil)
	repo.On("RecordAttendance", mock.Anything, "ev-1", "user-1", true).Return(nil)

	err := uc.RegisterForEvent(context.Background(), "ev-1", "user-1")
	assert.NoError(t, err)

	err = uc.CancelRegistration(context.Background(), "ev-1", "user-1")
	assert.NoError(t, err)

	err = uc.RecordAttendance(context.Background(), "ev-1", "user-1", true)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}
