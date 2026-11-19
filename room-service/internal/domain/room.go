package domain

import "context"

type Room struct {
	ID         string
	RoomNumber string
	Floor      int
	Capacity   int
}

type Resident struct {
	ID         string
	UserID     string
	RoomID     string
	RoomNumber string
	Role       string
	CheckInAt  int64
}

type DashboardStats struct {
	TotalResidents int
	TotalRooms     int
	OccupiedRooms  int
	AvailableBeds  int
}

type RoomRepository interface {
	CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (*Room, error)
	GetRoom(ctx context.Context, roomID string) (*Room, error)
	ListRooms(ctx context.Context, floor int) ([]*Room, error)
	
	AssignResident(ctx context.Context, userID, roomID string) (*Resident, error)
	RemoveResident(ctx context.Context, userID string) error
	GetResident(ctx context.Context, userID string) (*Resident, error)
	ListResidents(ctx context.Context, roomID, role string) ([]*Resident, error)
	UpdateResidentRole(ctx context.Context, userID, role string) error
	
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetResidentsByRoom(ctx context.Context, roomID string) ([]*Resident, error)
}

type RoomUseCase interface {
	CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (string, error)
	GetRoom(ctx context.Context, roomID string) (*Room, int, []*Resident, error)
	ListRooms(ctx context.Context, floor int) ([]*Room, error)
	
	AssignResident(ctx context.Context, userID, roomID string) (string, error)
	RemoveResident(ctx context.Context, userID string) error
	GetResident(ctx context.Context, userID string) (*Resident, error)
	ListResidents(ctx context.Context, roomID, role string) ([]*Resident, error)
	UpdateResidentRole(ctx context.Context, userID, role string) error
	
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
}
