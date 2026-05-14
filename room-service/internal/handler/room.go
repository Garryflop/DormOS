package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Garryflop/DormManage/room-service/internal/repository"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RoomHandler implements the gRPC RoomService interface manually
// (using plain structs until DormOS-gen-go is imported)
type RoomHandler struct {
	repo  *repository.RoomRepository
	rdb   *redis.Client
	nc    *nats.Conn
}

func NewRoomHandler(repo *repository.RoomRepository, rdb *redis.Client, nc *nats.Conn) *RoomHandler {
	return &RoomHandler{repo: repo, rdb: rdb, nc: nc}
}

// CreateRoom adds a new room (admin only, enforced at Gateway level)
func (h *RoomHandler) CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (string, error) {
	room, err := h.repo.CreateRoom(ctx, roomNumber, floor, capacity)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to create room: %v", err)
	}
	return room.ID, nil
}

// GetRoom returns room details including resident count
func (h *RoomHandler) GetRoom(ctx context.Context, roomID string) (*repository.Room, int, []*repository.Resident, error) {
	room, err := h.repo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, 0, nil, status.Errorf(codes.NotFound, "room not found: %v", err)
	}
	residents, err := h.repo.GetResidentsByRoom(ctx, roomID)
	if err != nil {
		return nil, 0, nil, status.Errorf(codes.Internal, "failed to get residents: %v", err)
	}
	return room, len(residents), residents, nil
}

// ListRooms returns all rooms, optionally filtered by floor
func (h *RoomHandler) ListRooms(ctx context.Context, floor int) ([]*repository.Room, error) {
	cacheKey := fmt.Sprintf("rooms:floor:%d", floor)
	// Try Redis cache first
	if cached, err := h.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var rooms []*repository.Room
		if json.Unmarshal(cached, &rooms) == nil {
			return rooms, nil
		}
	}

	rooms, err := h.repo.ListRooms(ctx, floor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list rooms: %v", err)
	}

	// Cache for 30 seconds
	if data, err := json.Marshal(rooms); err == nil {
		h.rdb.Set(ctx, cacheKey, data, 30*1e9)
	}
	return rooms, nil
}

// AssignResident assigns a user to a room and publishes NATS event
func (h *RoomHandler) AssignResident(ctx context.Context, userID, roomID string) (string, error) {
	resident, err := h.repo.AssignResident(ctx, userID, roomID)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to assign resident: %v", err)
	}

	// Publish NATS event for Notification Service
	if h.nc != nil {
		payload, _ := json.Marshal(map[string]string{
			"user_id": userID,
			"room_id": roomID,
			"event":   "resident.assigned",
		})
		_ = h.nc.Publish("resident.assigned", payload)
	}

	// Invalidate room cache
	h.rdb.Del(ctx, "rooms:floor:0")

	return resident.ID, nil
}

// RemoveResident removes a resident and publishes NATS event
func (h *RoomHandler) RemoveResident(ctx context.Context, userID string) error {
	if err := h.repo.RemoveResident(ctx, userID); err != nil {
		return status.Errorf(codes.Internal, "failed to remove resident: %v", err)
	}

	if h.nc != nil {
		payload, _ := json.Marshal(map[string]string{
			"user_id": userID,
			"event":   "resident.removed",
		})
		_ = h.nc.Publish("resident.removed", payload)
	}
	return nil
}

// GetResident returns a resident's info by user_id
func (h *RoomHandler) GetResident(ctx context.Context, userID string) (*repository.Resident, error) {
	res, err := h.repo.GetResident(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "resident not found: %v", err)
	}
	return res, nil
}

// ListResidents returns all residents, optionally filtered
func (h *RoomHandler) ListResidents(ctx context.Context, roomID, role string) ([]*repository.Resident, error) {
	list, err := h.repo.ListResidents(ctx, roomID, role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list residents: %v", err)
	}
	return list, nil
}

// UpdateResidentRole changes a resident's role
func (h *RoomHandler) UpdateResidentRole(ctx context.Context, userID, role string) error {
	validRoles := map[string]bool{"student": true, "manager": true, "admin": true}
	if !validRoles[role] {
		return status.Errorf(codes.InvalidArgument, "invalid role: %s", role)
	}
	if err := h.repo.UpdateResidentRole(ctx, userID, role); err != nil {
		return status.Errorf(codes.Internal, "failed to update role: %v", err)
	}
	return nil
}

// GetDashboardStats returns aggregated stats for the admin dashboard
func (h *RoomHandler) GetDashboardStats(ctx context.Context) (totalResidents, totalRooms, occupiedRooms, availableBeds int, err error) {
	cacheKey := "dashboard:stats"
	type stats struct {
		TotalResidents int `json:"total_residents"`
		TotalRooms     int `json:"total_rooms"`
		OccupiedRooms  int `json:"occupied_rooms"`
		AvailableBeds  int `json:"available_beds"`
	}

	if cached, err := h.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var s stats
		if json.Unmarshal(cached, &s) == nil {
			return s.TotalResidents, s.TotalRooms, s.OccupiedRooms, s.AvailableBeds, nil
		}
	}

	totalResidents, totalRooms, occupiedRooms, availableBeds, err = h.repo.GetDashboardStats(ctx)
	if err != nil {
		return 0, 0, 0, 0, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	s := stats{totalResidents, totalRooms, occupiedRooms, availableBeds}
	if data, jerr := json.Marshal(s); jerr == nil {
		h.rdb.Set(ctx, cacheKey, data, 60*1e9) // 60 seconds cache
	}
	return
}
