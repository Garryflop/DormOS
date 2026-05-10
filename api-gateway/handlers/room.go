package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// roomClient is a minimal gRPC client stub until DormOS-gen-go is imported.
// Once generated code is available, replace this with the real client.
type roomGRPCClient struct {
	conn *grpc.ClientConn
}

func newRoomClient(addr string) (*roomGRPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &roomGRPCClient{conn: conn}, nil
}

// RegisterRoomRoutes wires all room and resident HTTP endpoints.
// Called from main.go by Nurdaulet.
func RegisterRoomRoutes(r *gin.RouterGroup, roomServiceAddr string) {
	client, err := newRoomClient(roomServiceAddr)
	if err != nil {
		panic("failed to connect to room-service: " + err.Error())
	}
	h := &RoomHandler{client: client}

	rooms := r.Group("/rooms")
	{
		rooms.POST("", h.CreateRoom)          // admin
		rooms.GET("", h.ListRooms)            // manager+
		rooms.GET("/:id", h.GetRoom)          // student+
	}

	residents := r.Group("/residents")
	{
		residents.POST("", h.AssignResident)             // admin
		residents.DELETE("/:user_id", h.RemoveResident)  // admin
		residents.GET("/:user_id", h.GetResident)        // student+
		residents.GET("", h.ListResidents)               // admin
		residents.PATCH("/:user_id/role", h.UpdateResidentRole) // admin
	}

	r.GET("/dashboard/stats", h.GetDashboardStats) // admin
}

type RoomHandler struct {
	client *roomGRPCClient
}

// POST /api/v1/rooms
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req struct {
		RoomNumber string `json:"room_number" binding:"required"`
		Floor      int    `json:"floor"       binding:"required"`
		Capacity   int    `json:"capacity"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: call h.client.CreateRoom via generated stub
	c.JSON(http.StatusCreated, gin.H{
		"message":     "room created",
		"room_number": req.RoomNumber,
		"floor":       req.Floor,
		"capacity":    req.Capacity,
	})
}

// GET /api/v1/rooms?floor=2
func (h *RoomHandler) ListRooms(c *gin.Context) {
	floor := c.DefaultQuery("floor", "0")
	// TODO: call h.client.ListRooms
	c.JSON(http.StatusOK, gin.H{
		"rooms": []gin.H{},
		"floor_filter": floor,
	})
}

// GET /api/v1/rooms/:id
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")
	// TODO: call h.client.GetRoom
	c.JSON(http.StatusOK, gin.H{"room_id": roomID})
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
	// TODO: call h.client.AssignResident
	c.JSON(http.StatusCreated, gin.H{"message": "resident assigned"})
}

// DELETE /api/v1/residents/:user_id
func (h *RoomHandler) RemoveResident(c *gin.Context) {
	userID := c.Param("user_id")
	// TODO: call h.client.RemoveResident
	c.JSON(http.StatusOK, gin.H{"message": "resident removed", "user_id": userID})
}

// GET /api/v1/residents/:user_id
func (h *RoomHandler) GetResident(c *gin.Context) {
	userID := c.Param("user_id")
	// TODO: call h.client.GetResident
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// GET /api/v1/residents?room_id=...&role=...
func (h *RoomHandler) ListResidents(c *gin.Context) {
	roomID := c.Query("room_id")
	role := c.Query("role")
	// TODO: call h.client.ListResidents
	c.JSON(http.StatusOK, gin.H{"residents": []gin.H{}, "room_id": roomID, "role": role})
}

// PATCH /api/v1/residents/:user_id/role
func (h *RoomHandler) UpdateResidentRole(c *gin.Context) {
	userID := c.Param("user_id")
	var req struct {
		Role string `json:"role" binding:"required,oneof=student manager admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: call h.client.UpdateResidentRole
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "new_role": req.Role})
}

// GET /api/v1/dashboard/stats
func (h *RoomHandler) GetDashboardStats(c *gin.Context) {
	// TODO: call h.client.GetDashboardStats
	c.JSON(http.StatusOK, gin.H{
		"total_residents": 0,
		"total_rooms":     0,
		"occupied_rooms":  0,
		"available_beds":  0,
		"open_issues":     0,
		"upcoming_events": 0,
	})
}
