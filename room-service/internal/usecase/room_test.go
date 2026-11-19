package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Garryflop/DormManage/room-service/internal/domain"
	"github.com/Garryflop/DormManage/room-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock Repository ──────────────────────────────────────────────────────────

type mockRoomRepo struct{ mock.Mock }

func (m *mockRoomRepo) CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (*domain.Room, error) {
	args := m.Called(ctx, roomNumber, floor, capacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *mockRoomRepo) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *mockRoomRepo) ListRooms(ctx context.Context, floor int) ([]*domain.Room, error) {
	args := m.Called(ctx, floor)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func (m *mockRoomRepo) AssignResident(ctx context.Context, userID, roomID string) (*domain.Resident, error) {
	args := m.Called(ctx, userID, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resident), args.Error(1)
}

func (m *mockRoomRepo) RemoveResident(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockRoomRepo) GetResident(ctx context.Context, userID string) (*domain.Resident, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resident), args.Error(1)
}

func (m *mockRoomRepo) ListResidents(ctx context.Context, roomID, role string) ([]*domain.Resident, error) {
	args := m.Called(ctx, roomID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Resident), args.Error(1)
}

func (m *mockRoomRepo) UpdateResidentRole(ctx context.Context, userID, role string) error {
	return m.Called(ctx, userID, role).Error(0)
}

func (m *mockRoomRepo) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DashboardStats), args.Error(1)
}

func (m *mockRoomRepo) GetResidentsByRoom(ctx context.Context, roomID string) ([]*domain.Resident, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Resident), args.Error(1)
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func newUC(repo domain.RoomRepository) domain.RoomUseCase {
	// Pass nil for Redis and NATS — usecase handles nil gracefully
	return usecase.NewRoomUseCase(repo, nil, nil)
}

// ─── CreateRoom Tests ─────────────────────────────────────────────────────────

func TestCreateRoom_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("CreateRoom", mock.Anything, "305", 3, 2).
		Return(&domain.Room{ID: "room-1", RoomNumber: "305", Floor: 3, Capacity: 2}, nil)

	uc := newUC(repo)
	id, err := uc.CreateRoom(context.Background(), "305", 3, 2)

	assert.NoError(t, err)
	assert.Equal(t, "room-1", id)
	repo.AssertExpectations(t)
}

func TestCreateRoom_RepoError(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("CreateRoom", mock.Anything, "305", 3, 2).
		Return(nil, errors.New("duplicate room number"))

	uc := newUC(repo)
	_, err := uc.CreateRoom(context.Background(), "305", 3, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create room")
}

// ─── GetRoom Tests ────────────────────────────────────────────────────────────

func TestGetRoom_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	expectedRoom := &domain.Room{ID: "room-1", RoomNumber: "305", Floor: 3, Capacity: 4}
	residents := []*domain.Resident{
		{ID: "r1", UserID: "u1", RoomID: "room-1", Role: "student"},
		{ID: "r2", UserID: "u2", RoomID: "room-1", Role: "student"},
	}
	repo.On("GetRoom", mock.Anything, "room-1").Return(expectedRoom, nil)
	repo.On("GetResidentsByRoom", mock.Anything, "room-1").Return(residents, nil)

	uc := newUC(repo)
	room, occupied, resList, err := uc.GetRoom(context.Background(), "room-1")

	assert.NoError(t, err)
	assert.Equal(t, "305", room.RoomNumber)
	assert.Equal(t, 2, occupied)
	assert.Len(t, resList, 2)
	repo.AssertExpectations(t)
}

func TestGetRoom_NotFound(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("GetRoom", mock.Anything, "nonexistent").
		Return(nil, errors.New("room not found"))

	uc := newUC(repo)
	_, _, _, err := uc.GetRoom(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
}

// ─── AssignResident Tests ─────────────────────────────────────────────────────

func TestAssignResident_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	resident := &domain.Resident{
		ID: "res-1", UserID: "user-1", RoomID: "room-1", RoomNumber: "305", Role: "student",
	}
	repo.On("AssignResident", mock.Anything, "user-1", "room-1").Return(resident, nil)

	uc := newUC(repo)
	id, err := uc.AssignResident(context.Background(), "user-1", "room-1")

	assert.NoError(t, err)
	assert.Equal(t, "res-1", id)
	repo.AssertExpectations(t)
}

func TestAssignResident_Error(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("AssignResident", mock.Anything, "user-1", "room-1").
		Return(nil, errors.New("room full"))

	uc := newUC(repo)
	_, err := uc.AssignResident(context.Background(), "user-1", "room-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign resident")
}

// ─── RemoveResident Tests ─────────────────────────────────────────────────────

func TestRemoveResident_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("RemoveResident", mock.Anything, "user-1").Return(nil)

	uc := newUC(repo)
	err := uc.RemoveResident(context.Background(), "user-1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRemoveResident_Error(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("RemoveResident", mock.Anything, "user-1").
		Return(errors.New("not found"))

	uc := newUC(repo)
	err := uc.RemoveResident(context.Background(), "user-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove resident")
}

// ─── GetResident Tests ────────────────────────────────────────────────────────

func TestGetResident_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	expected := &domain.Resident{
		ID: "res-1", UserID: "user-1", RoomID: "room-1", RoomNumber: "305", Role: "student",
	}
	repo.On("GetResident", mock.Anything, "user-1").Return(expected, nil)

	uc := newUC(repo)
	result, err := uc.GetResident(context.Background(), "user-1")

	assert.NoError(t, err)
	assert.Equal(t, "res-1", result.ID)
	assert.Equal(t, "305", result.RoomNumber)
}

func TestGetResident_NotFound(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("GetResident", mock.Anything, "no-user").
		Return(nil, errors.New("not found"))

	uc := newUC(repo)
	_, err := uc.GetResident(context.Background(), "no-user")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resident not found")
}

// ─── ListResidents Tests ──────────────────────────────────────────────────────

func TestListResidents_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	residents := []*domain.Resident{
		{ID: "r1", UserID: "u1", Role: "student"},
		{ID: "r2", UserID: "u2", Role: "manager"},
	}
	repo.On("ListResidents", mock.Anything, "", "").Return(residents, nil)

	uc := newUC(repo)
	result, err := uc.ListResidents(context.Background(), "", "")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListResidents_Empty(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("ListResidents", mock.Anything, "room-99", "").
		Return([]*domain.Resident{}, nil)

	uc := newUC(repo)
	result, err := uc.ListResidents(context.Background(), "room-99", "")

	assert.NoError(t, err)
	assert.Empty(t, result)
}

// ─── UpdateResidentRole Tests ─────────────────────────────────────────────────

func TestUpdateResidentRole_ValidRoles(t *testing.T) {
	for _, role := range []string{"student", "manager", "admin"} {
		t.Run("valid_"+role, func(t *testing.T) {
			repo := &mockRoomRepo{}
			repo.On("UpdateResidentRole", mock.Anything, "user-1", role).Return(nil)

			uc := newUC(repo)
			err := uc.UpdateResidentRole(context.Background(), "user-1", role)

			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateResidentRole_InvalidRole(t *testing.T) {
	repo := &mockRoomRepo{}
	// The repo should never be called for an invalid role
	uc := newUC(repo)
	err := uc.UpdateResidentRole(context.Background(), "user-1", "superadmin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestUpdateResidentRole_EmptyRole(t *testing.T) {
	repo := &mockRoomRepo{}
	uc := newUC(repo)
	err := uc.UpdateResidentRole(context.Background(), "user-1", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

// ─── GetDashboardStats Tests ──────────────────────────────────────────────────

func TestGetDashboardStats_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	expected := &domain.DashboardStats{
		TotalResidents: 42,
		TotalRooms:     20,
		OccupiedRooms:  15,
		AvailableBeds:  18,
	}
	repo.On("GetDashboardStats", mock.Anything).Return(expected, nil)

	uc := newUC(repo)
	stats, err := uc.GetDashboardStats(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 42, stats.TotalResidents)
	assert.Equal(t, 20, stats.TotalRooms)
	assert.Equal(t, 15, stats.OccupiedRooms)
	assert.Equal(t, 18, stats.AvailableBeds)
}

func TestGetDashboardStats_Error(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("GetDashboardStats", mock.Anything).
		Return(nil, errors.New("db down"))

	uc := newUC(repo)
	_, err := uc.GetDashboardStats(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get stats")
}

// ─── ListRooms Tests (without Redis — nil client) ─────────────────────────────

func TestListRooms_NoCache_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	rooms := []*domain.Room{
		{ID: "r1", RoomNumber: "101", Floor: 1, Capacity: 2},
		{ID: "r2", RoomNumber: "102", Floor: 1, Capacity: 3},
	}
	repo.On("ListRooms", mock.Anything, 1).Return(rooms, nil)

	uc := newUC(repo)
	result, err := uc.ListRooms(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "101", result[0].RoomNumber)
}

func TestListRooms_RepoError(t *testing.T) {
	repo := &mockRoomRepo{}
	repo.On("ListRooms", mock.Anything, 5).
		Return(nil, errors.New("query failed"))

	uc := newUC(repo)
	_, err := uc.ListRooms(context.Background(), 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list rooms")
}
