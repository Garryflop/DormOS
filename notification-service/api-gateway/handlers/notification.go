package handlers

import (
	"context"
	"net/http"

	notificationv1 "github.com/Garryflop/DormOS-gen-go/notification/v1"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	client notificationv1.NotificationServiceClient
}

func NewNotificationHandler(
	client notificationv1.NotificationServiceClient,
) *NotificationHandler {

	return &NotificationHandler{
		client: client,
	}
}

func (h *NotificationHandler) RegisterNotificationRoutes(
	r *gin.RouterGroup,
) {

	notif := r.Group("/notifications")

	{
		notif.GET(
			"",
			h.ListNotifications,
		)

		notif.POST(
			"",
			h.SendNotification,
		)
	}
}

// ListNotifications
func (h *NotificationHandler) ListNotifications(
	c *gin.Context,
) {

	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "user_id is required",
			},
		)
		return
	}

	resp, err := h.client.ListNotifications(
		context.Background(),
		&notificationv1.ListNotificationsRequest{
			UserId: userID,
		},
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"notifications": resp.Notifications,
		},
	)
}

// SendNotification
func (h *NotificationHandler) SendNotification(
	c *gin.Context,
) {

	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Channel string `json:"channel"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	resp, err := h.client.SendNotification(
		context.Background(),
		&notificationv1.SendNotificationRequest{
			UserId:  req.UserID,
			Title:   req.Title,
			Message: req.Message,
			Channel: req.Channel,
		},
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": resp.Success,
		},
	)
}
