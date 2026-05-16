package grpc

import (
	"context"

	activityv1 "github.com/Garryflop/DormOS-gen-go/activity/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendNotification
func (h *Handler) SendNotification(
	ctx context.Context,
	req *activityv1.SendNotificationRequest,
) (*activityv1.SendNotificationResponse, error) {

	err := h.notificationUC.SendNotification(
		ctx,
		req.GetUserId(),
		"",
		req.GetTitle(),
		req.GetMessage(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"SendNotification: %v",
			err,
		)
	}

	return &activityv1.SendNotificationResponse{
		Success: true,
	}, nil
}

// ListMyNotifications
func (h *Handler) ListMyNotifications(
	ctx context.Context,
	req *activityv1.ListMyNotificationsRequest,
) (*activityv1.ListMyNotificationsResponse, error) {

	notifications, err := h.notificationUC.ListMyNotifications(
		ctx,
		req.GetUserId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"ListMyNotifications: %v",
			err,
		)
	}

	var items []*activityv1.NotificationItem

	for _, n := range notifications {

		items = append(items, &activityv1.NotificationItem{
			NotificationId: n.ID,
			Title:          n.Title,
			Message:        n.Message,
			IsRead:         n.IsRead,
			CreatedAt:      n.CreatedAt.Unix(),
		})
	}

	return &activityv1.ListMyNotificationsResponse{
		Notifications: items,
	}, nil
}

// MarkAsRead
func (h *Handler) MarkAsRead(
	ctx context.Context,
	req *activityv1.MarkAsReadRequest,
) (*activityv1.MarkAsReadResponse, error) {

	err := h.notificationUC.MarkAsRead(
		ctx,
		req.GetNotificationId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"MarkAsRead: %v",
			err,
		)
	}

	return &activityv1.MarkAsReadResponse{
		Success: true,
	}, nil
}

// MarkAllAsRead
func (h *Handler) MarkAllAsRead(
	ctx context.Context,
	req *activityv1.MarkAllAsReadRequest,
) (*activityv1.MarkAllAsReadResponse, error) {

	err := h.notificationUC.MarkAllAsRead(
		ctx,
		req.GetUserId(),
	)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"MarkAllAsRead: %v",
			err,
		)
	}

	return &activityv1.MarkAllAsReadResponse{}, nil
}
