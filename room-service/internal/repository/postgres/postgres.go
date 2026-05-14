package postgres

import (
	"context"
	"fmt"

	"github.com/Garryflop/DormManage/room-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) domain.RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateRoom(ctx context.Context, roomNumber string, floor, capacity int) (*domain.Room, error) {
	room := &domain.Room{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO rooms (room_number, floor, capacity) VALUES ($1, $2, $3) RETURNING id, room_number, floor, capacity`,
		roomNumber, floor, capacity,
	).Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Capacity)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return room, nil
}

func (r *RoomRepository) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	room := &domain.Room{}
	err := r.db.QueryRow(ctx,
		`SELECT id, room_number, floor, capacity FROM rooms WHERE id = $1`,
		roomID,
	).Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Capacity)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	return room, nil
}

func (r *RoomRepository) ListRooms(ctx context.Context, floor int) ([]*domain.Room, error) {
	query := `SELECT id, room_number, floor, capacity FROM rooms`
	args := []any{}
	if floor > 0 {
		query += ` WHERE floor = $1`
		args = append(args, floor)
	}
	query += ` ORDER BY floor, room_number`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		if err := rows.Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Capacity); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepository) AssignResident(ctx context.Context, userID, roomID string) (*domain.Resident, error) {
	resident := &domain.Resident{}
	var checkInAt int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO residents (user_id, room_id) VALUES ($1, $2)
		 RETURNING id, user_id, room_id, role, EXTRACT(EPOCH FROM check_in_at)::bigint`,
		userID, roomID,
	).Scan(&resident.ID, &resident.UserID, &resident.RoomID, &resident.Role, &checkInAt)
	if err != nil {
		return nil, fmt.Errorf("assign resident: %w", err)
	}
	resident.CheckInAt = checkInAt
	return resident, nil
}

func (r *RoomRepository) RemoveResident(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM residents WHERE user_id = $1`, userID)
	return err
}

func (r *RoomRepository) GetResident(ctx context.Context, userID string) (*domain.Resident, error) {
	res := &domain.Resident{}
	err := r.db.QueryRow(ctx,
		`SELECT r.id, r.user_id, r.room_id, ro.room_number, r.role, EXTRACT(EPOCH FROM r.check_in_at)::bigint
		 FROM residents r JOIN rooms ro ON ro.id = r.room_id
		 WHERE r.user_id = $1`,
		userID,
	).Scan(&res.ID, &res.UserID, &res.RoomID, &res.RoomNumber, &res.Role, &res.CheckInAt)
	if err != nil {
		return nil, fmt.Errorf("get resident: %w", err)
	}
	return res, nil
}

func (r *RoomRepository) ListResidents(ctx context.Context, roomID, role string) ([]*domain.Resident, error) {
	query := `SELECT r.id, r.user_id, r.room_id, ro.room_number, r.role, EXTRACT(EPOCH FROM r.check_in_at)::bigint
	          FROM residents r JOIN rooms ro ON ro.id = r.room_id WHERE 1=1`
	args := []any{}
	i := 1
	if roomID != "" {
		query += fmt.Sprintf(` AND r.room_id = $%d`, i)
		args = append(args, roomID)
		i++
	}
	if role != "" {
		query += fmt.Sprintf(` AND r.role = $%d`, i)
		args = append(args, role)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Resident
	for rows.Next() {
		res := &domain.Resident{}
		if err := rows.Scan(&res.ID, &res.UserID, &res.RoomID, &res.RoomNumber, &res.Role, &res.CheckInAt); err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, nil
}

func (r *RoomRepository) UpdateResidentRole(ctx context.Context, userID, role string) error {
	_, err := r.db.Exec(ctx, `UPDATE residents SET role = $1 WHERE user_id = $2`, role, userID)
	return err
}

func (r *RoomRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	var totalResidents, totalRooms, occupiedRooms int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM residents`).Scan(&totalResidents)
	if err != nil {
		return nil, err
	}
	var totalCapacity int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(capacity), 0) FROM rooms`).Scan(&totalRooms, &totalCapacity)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow(ctx, `SELECT COUNT(DISTINCT room_id) FROM residents`).Scan(&occupiedRooms)
	if err != nil {
		return nil, err
	}
	
	return &domain.DashboardStats{
		TotalResidents: totalResidents,
		TotalRooms:     totalRooms,
		OccupiedRooms:  occupiedRooms,
		AvailableBeds:  totalCapacity - totalResidents,
	}, nil
}

// GetResidentsByRoom returns all residents for a specific room
func (r *RoomRepository) GetResidentsByRoom(ctx context.Context, roomID string) ([]*domain.Resident, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, room_id, role, EXTRACT(EPOCH FROM check_in_at)::bigint FROM residents WHERE room_id = $1`,
		roomID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Resident
	for rows.Next() {
		res := &domain.Resident{}
		if err := rows.Scan(&res.ID, &res.UserID, &res.RoomID, &res.Role, &res.CheckInAt); err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, nil
}

// Ensure uuid is used
var _ = uuid.New
