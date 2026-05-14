package grpc

import (
	"context"

	"github.com/Garryflop/DormManage/room-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoomHandler struct {
	uc domain.RoomUseCase
}

func NewRoomHandler(uc domain.RoomUseCase) *RoomHandler {
	return &RoomHandler{uc: uc}
}

func (h *RoomHandler) CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (string, error) {
	return h.uc.CreateRoom(ctx, roomNumber, floor, capacity)
}

func (h *RoomHandler) GetRoom(ctx context.Context, roomID string) (*domain.Room, int, []*domain.Resident, error) {
	return h.uc.GetRoom(ctx, roomID)
}

func (h *RoomHandler) ListRooms(ctx context.Context, floor int) ([]*domain.Room, error) {
	return h.uc.ListRooms(ctx, floor)
}

func (h *RoomHandler) AssignResident(ctx context.Context, userID, roomID string) (string, error) {
	return h.uc.AssignResident(ctx, userID, roomID)
}

func (h *RoomHandler) RemoveResident(ctx context.Context, userID string) error {
	return h.uc.RemoveResident(ctx, userID)
}

func (h *RoomHandler) GetResident(ctx context.Context, userID string) (*domain.Resident, error) {
	return h.uc.GetResident(ctx, userID)
}

func (h *RoomHandler) ListResidents(ctx context.Context, roomID, role string) ([]*domain.Resident, error) {
	return h.uc.ListResidents(ctx, roomID, role)
}

func (h *RoomHandler) UpdateResidentRole(ctx context.Context, userID, role string) error {
	return h.uc.UpdateResidentRole(ctx, userID, role)
}

func (h *RoomHandler) GetDashboardStats(ctx context.Context) (int, int, int, int, error) {
	stats, err := h.uc.GetDashboardStats(ctx)
	if err != nil {
		return 0, 0, 0, 0, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}
	return stats.TotalResidents, stats.TotalRooms, stats.OccupiedRooms, stats.AvailableBeds, nil
}
