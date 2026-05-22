package handlers

import (
	"context"
	"net/http"

	activityv1 "github.com/Garryflop/DormOS-gen-go/activity/v1"
	notificationv1 "github.com/Garryflop/DormOS-gen-go/notification/v1"

	"github.com/gin-gonic/gin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ActivityHandler struct {
	client         notificationv1.NotificationServiceClient
	activityClient activityv1.ActivityServiceClient
}

func NewActivityHandler(
	notificationServiceAddr string,
) (*ActivityHandler, error) {

	conn, err := grpc.Dial(
		notificationServiceAddr,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	return &ActivityHandler{
		client:         notificationv1.NewNotificationServiceClient(conn),
		activityClient: activityv1.NewActivityServiceClient(conn),
	}, nil
}

func (h *ActivityHandler) RegisterActivityRoutes(
	r *gin.RouterGroup,
) {
	// Notifications
	notifications := r.Group("/notifications")
	{
		notifications.GET("", h.ListNotifications)
		notifications.POST("", h.SendNotification)
		notifications.PATCH("/:id/read", h.MarkAsRead)
		notifications.POST("/read-all", h.MarkAllAsRead)
	}

	// Events
	events := r.Group("/events")
	{
		events.GET("", h.ListEvents)
		events.POST("", h.CreateEvent)
		events.PUT("/:id", h.UpdateEvent)
		events.DELETE("/:id", h.DeleteEvent)
		events.POST("/:id/register", h.RegisterForEvent)
		events.POST("/:id/cancel", h.CancelRegistration)
	}
}

// SendNotification
func (h *ActivityHandler) SendNotification(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Channel string `json:"channel"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.activityClient.SendNotification(
		context.Background(),
		&activityv1.SendNotificationRequest{
			UserId:  req.UserID,
			Title:   req.Title,
			Message: req.Message,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// ListNotifications
func (h *ActivityHandler) ListNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	resp, err := h.activityClient.ListMyNotifications(
		context.Background(),
		&activityv1.ListMyNotificationsRequest{
			UserId: userID,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": resp.Notifications})
}

// MarkAsRead
func (h *ActivityHandler) MarkAsRead(c *gin.Context) {
	notifID := c.Param("id")
	resp, err := h.activityClient.MarkAsRead(
		context.Background(),
		&activityv1.MarkAsReadRequest{
			NotificationId: notifID,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// MarkAllAsRead
func (h *ActivityHandler) MarkAllAsRead(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.activityClient.MarkAllAsRead(
		context.Background(),
		&activityv1.MarkAllAsReadRequest{
			UserId: req.UserID,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated_count": resp.UpdatedCount})
}

// CreateEvent
func (h *ActivityHandler) CreateEvent(c *gin.Context) {
	var req struct {
		Title           string `json:"title" binding:"required"`
		Description     string `json:"description" binding:"required"`
		Location        string `json:"location" binding:"required"`
		StartTime       int64  `json:"start_time" binding:"required"`
		EndTime         int64  `json:"end_time" binding:"required"`
		MaxParticipants int32  `json:"max_participants" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.activityClient.CreateEvent(
		c.Request.Context(),
		&activityv1.CreateEventRequest{
			Title:           req.Title,
			Description:     req.Description,
			Location:        req.Location,
			StartTime:       req.StartTime,
			EndTime:         req.EndTime,
			MaxParticipants: req.MaxParticipants,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event_id": resp.EventId})
}

// ListEvents
func (h *ActivityHandler) ListEvents(c *gin.Context) {
	resp, err := h.activityClient.ListEvents(
		c.Request.Context(),
		&activityv1.ListEventsRequest{},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": resp.Events})
}

// UpdateEvent
func (h *ActivityHandler) UpdateEvent(c *gin.Context) {
	var req struct {
		Title           string `json:"title" binding:"required"`
		Description     string `json:"description" binding:"required"`
		Location        string `json:"location" binding:"required"`
		StartTime       int64  `json:"start_time" binding:"required"`
		EndTime         int64  `json:"end_time" binding:"required"`
		MaxParticipants int32  `json:"max_participants" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.activityClient.UpdateEvent(
		c.Request.Context(),
		&activityv1.UpdateEventRequest{
			EventId:         c.Param("id"),
			Title:           req.Title,
			Description:     req.Description,
			Location:        req.Location,
			StartTime:       req.StartTime,
			EndTime:         req.EndTime,
			MaxParticipants: req.MaxParticipants,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// DeleteEvent
func (h *ActivityHandler) DeleteEvent(c *gin.Context) {
	resp, err := h.activityClient.DeleteEvent(
		c.Request.Context(),
		&activityv1.DeleteEventRequest{
			EventId: c.Param("id"),
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// RegisterForEvent
func (h *ActivityHandler) RegisterForEvent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.activityClient.RegisterForEvent(
		c.Request.Context(),
		&activityv1.RegisterForEventRequest{
			EventId: c.Param("id"),
			UserId:  userID.(string),
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// CancelRegistration
func (h *ActivityHandler) CancelRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.activityClient.CancelRegistration(
		c.Request.Context(),
		&activityv1.CancelRegistrationRequest{
			EventId: c.Param("id"),
			UserId:  userID.(string),
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}
