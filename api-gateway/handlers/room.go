package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	roomv1 "github.com/Garryflop/DormOS-gen-go/room/v1"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RoomHandler struct {
	client roomv1.RoomServiceClient
}

// RegisterRoomRoutes wires all room and resident HTTP endpoints.
func RegisterRoomRoutes(r *gin.RouterGroup, roomServiceAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, roomServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		panic("failed to connect to room-service: " + err.Error())
	}

	h := &RoomHandler{client: roomv1.NewRoomServiceClient(conn)}

	rooms := r.Group("/rooms")
	{
		rooms.POST("", h.CreateRoom)
		rooms.GET("", h.ListRooms)
		rooms.GET("/:id", h.GetRoom)
	}

	residents := r.Group("/residents")
	{
		residents.POST("", h.AssignResident)
		residents.DELETE("/:user_id", h.RemoveResident)
		residents.GET("/:user_id", h.GetResident)
		residents.GET("", h.ListResidents)
		residents.PATCH("/:user_id/role", h.UpdateResidentRole)
	}

	r.GET("/dashboard/stats", h.GetDashboardStats)
}

// POST /api/v1/rooms
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req struct {
		RoomNumber string `json:"room_number" binding:"required"`
		Floor      int32  `json:"floor"       binding:"required"`
		Capacity   int32  `json:"capacity"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CreateRoom(c.Request.Context(), &roomv1.CreateRoomRequest{
		RoomNumber: req.RoomNumber,
		Floor:      req.Floor,
		Capacity:   req.Capacity,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"room_id": resp.RoomId})
}

// GET /api/v1/rooms?floor=2
func (h *RoomHandler) ListRooms(c *gin.Context) {
	floorStr := c.DefaultQuery("floor", "0")
	floor, _ := strconv.ParseInt(floorStr, 10, 32)
	
	resp, err := h.client.ListRooms(c.Request.Context(), &roomv1.ListRoomsRequest{
		Floor: int32(floor),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rooms": resp.Rooms})
}

// GET /api/v1/rooms/:id
func (h *RoomHandler) GetRoom(c *gin.Context) {
	resp, err := h.client.GetRoom(c.Request.Context(), &roomv1.GetRoomRequest{
		RoomId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"room_id":     resp.RoomId,
		"room_number": resp.RoomNumber,
		"floor":       resp.Floor,
		"capacity":    resp.Capacity,
		"occupied":    resp.Occupied,
		"residents":   resp.Residents,
	})
}

// POST /api/v1/residents
func (h *RoomHandler) AssignResident(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		RoomID string `json:"room_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.AssignResident(c.Request.Context(), &roomv1.AssignResidentRequest{
		UserId: req.UserID,
		RoomId: req.RoomID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"resident_id": resp.ResidentId})
}

// DELETE /api/v1/residents/:user_id
func (h *RoomHandler) RemoveResident(c *gin.Context) {
	_, err := h.client.RemoveResident(c.Request.Context(), &roomv1.RemoveResidentRequest{
		UserId: c.Param("user_id"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "resident removed"})
}

// GET /api/v1/residents/:user_id
func (h *RoomHandler) GetResident(c *gin.Context) {
	resp, err := h.client.GetResident(c.Request.Context(), &roomv1.GetResidentRequest{
		UserId: c.Param("user_id"),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"resident_id":   resp.ResidentId,
		"user_id":       resp.UserId,
		"room_number":   resp.RoomNumber,
		"role":          resp.Role,
		"check_in_date": resp.CheckInDate,
	})
}

// GET /api/v1/residents?room_id=...&role=...
func (h *RoomHandler) ListResidents(c *gin.Context) {
	resp, err := h.client.ListResidents(c.Request.Context(), &roomv1.ListResidentsRequest{
		RoomId: c.Query("room_id"),
		Role:   c.Query("role"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"residents": resp.Residents})
}

// PATCH /api/v1/residents/:user_id/role
func (h *RoomHandler) UpdateResidentRole(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required,oneof=student manager admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.client.UpdateResidentRole(c.Request.Context(), &roomv1.UpdateResidentRoleRequest{
		UserId:  c.Param("user_id"),
		NewRole: req.Role,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// GET /api/v1/dashboard/stats
func (h *RoomHandler) GetDashboardStats(c *gin.Context) {
	resp, err := h.client.GetDashboardStats(c.Request.Context(), &roomv1.GetDashboardStatsRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total_residents": resp.TotalResidents,
		"total_rooms":     resp.TotalRooms,
		"occupied_rooms":  resp.OccupiedRooms,
		"available_beds":  resp.AvailableBeds,
	})
}
