package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Garryflop/DormManage/room-service/internal/domain"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type roomUseCase struct {
	repo domain.RoomRepository
	rdb  *redis.Client
	nc   *nats.Conn
}

func NewRoomUseCase(repo domain.RoomRepository, rdb *redis.Client, nc *nats.Conn) domain.RoomUseCase {
	return &roomUseCase{
		repo: repo,
		rdb:  rdb,
		nc:   nc,
	}
}

func (uc *roomUseCase) CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (string, error) {
	room, err := uc.repo.CreateRoom(ctx, roomNumber, floor, capacity)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to create room: %v", err)
	}
	return room.ID, nil
}

func (uc *roomUseCase) GetRoom(ctx context.Context, roomID string) (*domain.Room, int, []*domain.Resident, error) {
	room, err := uc.repo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, 0, nil, status.Errorf(codes.NotFound, "room not found: %v", err)
	}
	residents, err := uc.repo.GetResidentsByRoom(ctx, roomID)
	if err != nil {
		return nil, 0, nil, status.Errorf(codes.Internal, "failed to get residents: %v", err)
	}
	return room, len(residents), residents, nil
}

func (uc *roomUseCase) ListRooms(ctx context.Context, floor int) ([]*domain.Room, error) {
	cacheKey := fmt.Sprintf("rooms:floor:%d", floor)
	
	// Try Redis cache first
	if cached, err := uc.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var rooms []*domain.Room
		if json.Unmarshal(cached, &rooms) == nil {
			return rooms, nil
		}
	}

	rooms, err := uc.repo.ListRooms(ctx, floor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list rooms: %v", err)
	}

	// Cache for 30 seconds
	if data, err := json.Marshal(rooms); err == nil {
		uc.rdb.Set(ctx, cacheKey, data, 30*1e9)
	}
	return rooms, nil
}

func (uc *roomUseCase) AssignResident(ctx context.Context, userID, roomID string) (string, error) {
	resident, err := uc.repo.AssignResident(ctx, userID, roomID)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to assign resident: %v", err)
	}

	// Publish NATS event
	if uc.nc != nil {
		payload, _ := json.Marshal(map[string]string{
			"user_id": userID,
			"room_id": roomID,
			"event":   "resident.assigned",
		})
		_ = uc.nc.Publish("resident.assigned", payload)
	}

	// Invalidate room cache
	uc.rdb.Del(ctx, "rooms:floor:0")

	return resident.ID, nil
}

func (uc *roomUseCase) RemoveResident(ctx context.Context, userID string) error {
	if err := uc.repo.RemoveResident(ctx, userID); err != nil {
		return status.Errorf(codes.Internal, "failed to remove resident: %v", err)
	}

	if uc.nc != nil {
		payload, _ := json.Marshal(map[string]string{
			"user_id": userID,
			"event":   "resident.removed",
		})
		_ = uc.nc.Publish("resident.removed", payload)
	}
	return nil
}

func (uc *roomUseCase) GetResident(ctx context.Context, userID string) (*domain.Resident, error) {
	res, err := uc.repo.GetResident(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "resident not found: %v", err)
	}
	return res, nil
}

func (uc *roomUseCase) ListResidents(ctx context.Context, roomID, role string) ([]*domain.Resident, error) {
	list, err := uc.repo.ListResidents(ctx, roomID, role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list residents: %v", err)
	}
	return list, nil
}

func (uc *roomUseCase) UpdateResidentRole(ctx context.Context, userID, role string) error {
	validRoles := map[string]bool{"student": true, "manager": true, "admin": true}
	if !validRoles[role] {
		return status.Errorf(codes.InvalidArgument, "invalid role: %s", role)
	}
	if err := uc.repo.UpdateResidentRole(ctx, userID, role); err != nil {
		return status.Errorf(codes.Internal, "failed to update role: %v", err)
	}
	return nil
}

func (uc *roomUseCase) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	cacheKey := "dashboard:stats"

	if cached, err := uc.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var s domain.DashboardStats
		if json.Unmarshal(cached, &s) == nil {
			return &s, nil
		}
	}

	stats, err := uc.repo.GetDashboardStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	if data, jerr := json.Marshal(stats); jerr == nil {
		uc.rdb.Set(ctx, cacheKey, data, 60*1e9)
	}
	return stats, nil
}
