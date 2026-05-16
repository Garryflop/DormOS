package domain

import "time"

type Event struct {
	ID              string
	Title           string
	Description     string
	Location        string
	EventDate       time.Time
	MaxParticipants int32
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type EventRegistration struct {
	ID           string
	EventID      string
	UserID       string
	RegisteredAt time.Time
	Attended     bool
}
