package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dormos/notification-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	eventsListCacheKey = "events:list"
	eventsCacheTTL     = 5 * time.Minute
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) (*domain.Event, error)
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	List(ctx context.Context) ([]*domain.Event, error)
	Update(ctx context.Context, event *domain.Event) (*domain.Event, error)
	Delete(ctx context.Context, id string) error

	RegisterUser(ctx context.Context, eventID, userID string) error
	CancelRegistration(ctx context.Context, eventID, userID string) error
	RecordAttendance(ctx context.Context, eventID, userID string, attended bool) error

	GetRegisteredUserIDs(ctx context.Context, eventID string) ([]string, error)
}

type NATSPublisher interface {
	Publish(subject string, data []byte) error
}

type ActivityUseCase struct {
	repo      EventRepository
	redis     *redis.Client
	publisher NATSPublisher
}

func NewActivityUseCase(
	repo EventRepository,
	redisClient *redis.Client,
	publisher NATSPublisher,
) *ActivityUseCase {
	return &ActivityUseCase{
		repo:      repo,
		redis:     redisClient,
		publisher: publisher,
	}
}

func (uc *ActivityUseCase) CreateEvent(
	ctx context.Context,
	event *domain.Event,
) (*domain.Event, error) {

	created, err := uc.repo.Create(ctx, event)
	if err != nil {
		return nil, fmt.Errorf(
			"ActivityUseCase.CreateEvent: %w",
			err,
		)
	}

	_ = uc.redis.Del(ctx, eventsListCacheKey).Err()

	payload := domain.NATSEventPayload{
		EventID: created.ID,
		Title:   fmt.Sprintf("Новое мероприятие: %s", created.Title),
		Message: fmt.Sprintf(
			"Новое мероприятие '%s' запланировано на %s в %s",
			created.Title,
			created.EventDate.Format("02.01.2006 15:04"),
			created.Location,
		),
	}

	data, _ := json.Marshal(payload)

	_ = uc.publisher.Publish(
		"event.created",
		data,
	)

	return created, nil
}

func (uc *ActivityUseCase) GetEvent(
	ctx context.Context,
	id string,
) (*domain.Event, error) {

	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(
			"ActivityUseCase.GetEvent: %w",
			err,
		)
	}

	return event, nil
}

func (uc *ActivityUseCase) ListEvents(
	ctx context.Context,
) ([]*domain.Event, error) {

	cached, err := uc.redis.Get(
		ctx,
		eventsListCacheKey,
	).Bytes()

	if err == nil {
		var events []*domain.Event

		if json.Unmarshal(cached, &events) == nil {
			return events, nil
		}
	}

	events, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"ActivityUseCase.ListEvents: %w",
			err,
		)
	}

	data, err := json.Marshal(events)
	if err == nil {
		_ = uc.redis.Set(
			ctx,
			eventsListCacheKey,
			data,
			eventsCacheTTL,
		).Err()
	}

	return events, nil
}

func (uc *ActivityUseCase) UpdateEvent(
	ctx context.Context,
	event *domain.Event,
) (*domain.Event, error) {

	updated, err := uc.repo.Update(ctx, event)
	if err != nil {
		return nil, fmt.Errorf(
			"ActivityUseCase.UpdateEvent: %w",
			err,
		)
	}

	_ = uc.redis.Del(ctx, eventsListCacheKey).Err()

	return updated, nil
}

func (uc *ActivityUseCase) DeleteEvent(
	ctx context.Context,
	id string,
) error {

	userIDs, _ := uc.repo.GetRegisteredUserIDs(
		ctx,
		id,
	)

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf(
			"ActivityUseCase.DeleteEvent: %w",
			err,
		)
	}

	_ = uc.redis.Del(ctx, eventsListCacheKey).Err()

	payload := domain.NATSEventPayload{
		EventID: id,
		UserIDs: userIDs,
		Title:   "Мероприятие отменено",
		Message: "К сожалению, мероприятие было отменено.",
	}

	data, _ := json.Marshal(payload)

	_ = uc.publisher.Publish(
		"event.cancelled",
		data,
	)

	return nil
}

func (uc *ActivityUseCase) RegisterForEvent(
	ctx context.Context,
	eventID string,
	userID string,
) error {

	err := uc.repo.RegisterUser(
		ctx,
		eventID,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"ActivityUseCase.RegisterForEvent: %w",
			err,
		)
	}

	return nil
}

func (uc *ActivityUseCase) CancelRegistration(
	ctx context.Context,
	eventID string,
	userID string,
) error {

	err := uc.repo.CancelRegistration(
		ctx,
		eventID,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"ActivityUseCase.CancelRegistration: %w",
			err,
		)
	}

	return nil
}

func (uc *ActivityUseCase) RecordAttendance(
	ctx context.Context,
	eventID string,
	userID string,
	attended bool,
) error {

	err := uc.repo.RecordAttendance(
		ctx,
		eventID,
		userID,
		attended,
	)

	if err != nil {
		return fmt.Errorf(
			"ActivityUseCase.RecordAttendance: %w",
			err,
		)
	}

	return nil
}
