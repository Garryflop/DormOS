package grpc

import (
	"github.com/dormos/notification-service/internal/usecase"

	activityv1 "github.com/Garryflop/DormOS-gen-go/activity/v1"
)

type Handler struct {
	activityv1.UnimplementedActivityServiceServer

	activityUC     *usecase.ActivityUseCase
	notificationUC *usecase.NotificationUseCase
}

func NewHandler(
	activityUC *usecase.ActivityUseCase,
	notifUC *usecase.NotificationUseCase,
) *Handler {

	return &Handler{
		activityUC:     activityUC,
		notificationUC: notifUC,
	}
}
