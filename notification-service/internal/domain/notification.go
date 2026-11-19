package domain

import "time"

type Notification struct {
	ID        string
	UserID    string
	Title     string
	Message   string
	IsRead    bool
	CreatedAt time.Time
}

type NATSEventPayload struct {
	UserID  string   `json:"user_id,omitempty"`
	UserIDs []string `json:"user_ids,omitempty"`
	Email   string   `json:"email,omitempty"`
	Emails  []string `json:"emails,omitempty"`
	Title   string   `json:"title"`
	Message string   `json:"message"`
	EventID string   `json:"event_id,omitempty"`
	RoomNo  string   `json:"room_no,omitempty"`
	Code    string   `json:"code,omitempty"`
}
