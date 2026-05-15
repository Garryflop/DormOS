package postgres

import (
	"context"

	"github.com/dormos/notification-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(
	db *pgxpool.Pool,
) *EventRepository {

	return &EventRepository{
		db: db,
	}
}

func (r *EventRepository) Create(
	ctx context.Context,
	event *domain.Event,
) (*domain.Event, error) {

	query := `
		INSERT INTO events
		(
			title,
			description,
			location,
			event_date,
			max_participants,
			created_by
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING
			id,
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		event.Title,
		event.Description,
		event.Location,
		event.EventDate,
		event.MaxParticipants,
		event.CreatedBy,
	).Scan(
		&event.ID,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Event, error) {

	query := `
		SELECT
			id,
			title,
			description,
			location,
			event_date,
			max_participants,
			created_by,
			created_at,
			updated_at
		FROM events
		WHERE id = $1
	`

	var event domain.Event

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.Location,
		&event.EventDate,
		&event.MaxParticipants,
		&event.CreatedBy,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *EventRepository) List(
	ctx context.Context,
) ([]*domain.Event, error) {

	query := `
		SELECT
			id,
			title,
			description,
			location,
			event_date,
			max_participants,
			created_by,
			created_at,
			updated_at
		FROM events
		ORDER BY event_date ASC
	`

	rows, err := r.db.Query(
		ctx,
		query,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []*domain.Event

	for rows.Next() {

		var event domain.Event

		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.Location,
			&event.EventDate,
			&event.MaxParticipants,
			&event.CreatedBy,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		events = append(
			events,
			&event,
		)
	}

	return events, nil
}

func (r *EventRepository) Update(
	ctx context.Context,
	event *domain.Event,
) (*domain.Event, error) {

	query := `
		UPDATE events
		SET
			title = $1,
			description = $2,
			location = $3,
			event_date = $4,
			max_participants = $5,
			updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		event.Title,
		event.Description,
		event.Location,
		event.EventDate,
		event.MaxParticipants,
		event.ID,
	).Scan(
		&event.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) Delete(
	ctx context.Context,
	id string,
) error {

	query := `
		DELETE FROM events
		WHERE id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	return err
}

func (r *EventRepository) RegisterUser(
	ctx context.Context,
	eventID string,
	userID string,
) error {

	query := `
		INSERT INTO event_registrations
		(
			event_id,
			user_id
		)
		VALUES ($1,$2)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		eventID,
		userID,
	)

	return err
}

func (r *EventRepository) CancelRegistration(
	ctx context.Context,
	eventID string,
	userID string,
) error {

	query := `
		DELETE FROM event_registrations
		WHERE event_id = $1
		AND user_id = $2
	`

	_, err := r.db.Exec(
		ctx,
		query,
		eventID,
		userID,
	)

	return err
}

func (r *EventRepository) RecordAttendance(
	ctx context.Context,
	eventID string,
	userID string,
	attended bool,
) error {

	query := `
		UPDATE event_registrations
		SET attended = $1
		WHERE event_id = $2
		AND user_id = $3
	`

	_, err := r.db.Exec(
		ctx,
		query,
		attended,
		eventID,
		userID,
	)

	return err
}

func (r *EventRepository) GetRegisteredUserIDs(
	ctx context.Context,
	eventID string,
) ([]string, error) {

	query := `
		SELECT user_id
		FROM event_registrations
		WHERE event_id = $1
	`

	rows, err := r.db.Query(
		ctx,
		query,
		eventID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var userIDs []string

	for rows.Next() {

		var userID string

		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}

		userIDs = append(
			userIDs,
			userID,
		)
	}

	return userIDs, nil
}
