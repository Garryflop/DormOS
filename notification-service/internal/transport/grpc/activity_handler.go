package grpc

import (
	"context"
	"strings"
	"time"

	"github.com/dormos/notification-service/internal/domain"

	activityv1 "github.com/Garryflop/DormOS-gen-go/activity/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateEvent
func (h *Handler) CreateEvent(
	ctx context.Context,
	req *activityv1.CreateEventRequest,
) (*activityv1.CreateEventResponse, error) {

	event := &domain.Event{
		Title:           req.GetTitle(),
		Description:     req.GetDescription(),
		Location:        req.GetLocation(),
		EventDate:       time.Unix(req.GetStartTime(), 0),
		MaxParticipants: req.GetMaxParticipants(),
		CreatedBy:       "550e8400-e29b-41d4-a716-446655440000",
	}

	created, err := h.activityUC.CreateEvent(
		ctx,
		event,
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"CreateEvent: %v",
			err,
		)
	}

	return &activityv1.CreateEventResponse{
		EventId: created.ID,
	}, nil
}

// GetEvent
func (h *Handler) GetEvent(
	ctx context.Context,
	req *activityv1.GetEventRequest,
) (*activityv1.GetEventResponse, error) {

	event, err := h.activityUC.GetEvent(
		ctx,
		req.GetEventId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"GetEvent: %v",
			err,
		)
	}

	userIDs, _ := h.activityUC.GetRegisteredUserIDs(ctx, event.ID)
	desc := event.Description
	if len(userIDs) > 0 {
		desc = desc + "\n\n[participants:" + strings.Join(userIDs, ",") + "]"
	} else {
		desc = desc + "\n\n[participants:]"
	}

	return &activityv1.GetEventResponse{
		EventId:         event.ID,
		Title:           event.Title,
		Description:     desc,
		Location:        event.Location,
		StartTime:       event.EventDate.Unix(),
		EndTime:         event.EventDate.Unix(),
		MaxParticipants: event.MaxParticipants,
		RegisteredCount: int32(len(userIDs)),
		CreatedAt:       event.CreatedAt.Unix(),
	}, nil
}

// ListEvents
func (h *Handler) ListEvents(
	ctx context.Context,
	req *activityv1.ListEventsRequest,
) (*activityv1.ListEventsResponse, error) {

	events, err := h.activityUC.ListEvents(ctx)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"ListEvents: %v",
			err,
		)
	}

	var items []*activityv1.GetEventResponse

	for _, e := range events {
		userIDs, _ := h.activityUC.GetRegisteredUserIDs(ctx, e.ID)
		desc := e.Description
		if len(userIDs) > 0 {
			desc = desc + "\n\n[participants:" + strings.Join(userIDs, ",") + "]"
		} else {
			desc = desc + "\n\n[participants:]"
		}

		items = append(items, &activityv1.GetEventResponse{
			EventId:         e.ID,
			Title:           e.Title,
			Description:     desc,
			Location:        e.Location,
			StartTime:       e.EventDate.Unix(),
			EndTime:         e.EventDate.Unix(),
			MaxParticipants: e.MaxParticipants,
			RegisteredCount: int32(len(userIDs)),
			CreatedAt:       e.CreatedAt.Unix(),
		})
	}

	return &activityv1.ListEventsResponse{
		Events: items,
	}, nil
}

// UpdateEvent
func (h *Handler) UpdateEvent(
	ctx context.Context,
	req *activityv1.UpdateEventRequest,
) (*activityv1.UpdateEventResponse, error) {

	event := &domain.Event{
		ID:              req.GetEventId(),
		Title:           req.GetTitle(),
		Description:     req.GetDescription(),
		Location:        req.GetLocation(),
		EventDate:       time.Unix(req.GetStartTime(), 0),
		MaxParticipants: req.GetMaxParticipants(),
	}

	_, err := h.activityUC.UpdateEvent(
		ctx,
		event,
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"UpdateEvent: %v",
			err,
		)
	}

	return &activityv1.UpdateEventResponse{
		Success: true,
	}, nil
}

// DeleteEvent
func (h *Handler) DeleteEvent(
	ctx context.Context,
	req *activityv1.DeleteEventRequest,
) (*activityv1.DeleteEventResponse, error) {

	err := h.activityUC.DeleteEvent(
		ctx,
		req.GetEventId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"DeleteEvent: %v",
			err,
		)
	}

	return &activityv1.DeleteEventResponse{
		Success: true,
	}, nil
}

// RegisterForEvent
func (h *Handler) RegisterForEvent(
	ctx context.Context,
	req *activityv1.RegisterForEventRequest,
) (*activityv1.RegisterForEventResponse, error) {

	err := h.activityUC.RegisterForEvent(
		ctx,
		req.GetEventId(),
		req.GetUserId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"RegisterForEvent: %v",
			err,
		)
	}

	return &activityv1.RegisterForEventResponse{
		Success: true,
	}, nil
}

// CancelRegistration
func (h *Handler) CancelRegistration(
	ctx context.Context,
	req *activityv1.CancelRegistrationRequest,
) (*activityv1.CancelRegistrationResponse, error) {

	err := h.activityUC.CancelRegistration(
		ctx,
		req.GetEventId(),
		req.GetUserId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"CancelRegistration: %v",
			err,
		)
	}

	return &activityv1.CancelRegistrationResponse{
		Success: true,
	}, nil
}

// RecordAttendance
func (h *Handler) RecordAttendance(
	ctx context.Context,
	req *activityv1.RecordAttendanceRequest,
) (*activityv1.RecordAttendanceResponse, error) {

	err := h.activityUC.RecordAttendance(
		ctx,
		req.GetEventId(),
		req.GetUserId(),
		req.GetAttended(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"RecordAttendance: %v",
			err,
		)
	}

	return &activityv1.RecordAttendanceResponse{
		Success: true,
	}, nil
}
