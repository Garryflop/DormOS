package grpc

import (
	"context"
	"strings"

	"github.com/Garryflop/DormManage/room-service/internal/domain"
	roomv1 "github.com/Garryflop/DormOS-gen-go/room/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoomHandler struct {
	roomv1.UnimplementedRoomServiceServer
	uc domain.RoomUseCase
}

func NewRoomHandler(uc domain.RoomUseCase) *RoomHandler {
	return &RoomHandler{uc: uc}
}

func (h *RoomHandler) CreateRoom(ctx context.Context, req *roomv1.CreateRoomRequest) (*roomv1.CreateRoomResponse, error) {
	id, err := h.uc.CreateRoom(ctx, req.RoomNumber, int(req.Floor), int(req.Capacity))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create room: %v", err)
	}
	return &roomv1.CreateRoomResponse{RoomId: id}, nil
}

func (h *RoomHandler) GetRoom(ctx context.Context, req *roomv1.GetRoomRequest) (*roomv1.GetRoomResponse, error) {
	room, occupied, residents, err := h.uc.GetRoom(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "room not found: %v", err)
	}
	
	resInfos := make([]*roomv1.ResidentInfo, len(residents))
	for i, r := range residents {
		resInfos[i] = &roomv1.ResidentInfo{
			UserId:   r.UserID,
			FullName: "", // We don't store full name in room-service
			Role:     r.Role,
		}
	}

	return &roomv1.GetRoomResponse{
		RoomId:     room.ID,
		RoomNumber: room.RoomNumber,
		Floor:      int32(room.Floor),
		Capacity:   int32(room.Capacity),
		Occupied:   int32(occupied),
		Residents:  resInfos,
	}, nil
}

func (h *RoomHandler) ListRooms(ctx context.Context, req *roomv1.ListRoomsRequest) (*roomv1.ListRoomsResponse, error) {
	rooms, err := h.uc.ListRooms(ctx, int(req.Floor))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list rooms: %v", err)
	}
	
	// Fetch all residents to associate details
	residents, _ := h.uc.ListResidents(ctx, "", "")
	roomResidents := make(map[string][]*roomv1.ResidentInfo)
	if residents != nil {
		for _, res := range residents {
			roomNum := res.RoomID
			if roomNum != "" && roomNum != "Unassigned" {
				parts := strings.Split(res.UserID, ":")
				resUUID := res.UserID
				resName := "Registered Student"
				if len(parts) > 1 {
					resUUID = parts[0]
					resName = parts[1]
				}
				roomResidents[roomNum] = append(roomResidents[roomNum], &roomv1.ResidentInfo{
					UserId:   resUUID,
					FullName: resName,
					Role:     res.Role,
				})
			}
		}
	}
	
	respRooms := make([]*roomv1.GetRoomResponse, len(rooms))
	for i, r := range rooms {
		resInfos := roomResidents[r.RoomNumber]
		occupiedCount := int32(len(resInfos))
		
		respRooms[i] = &roomv1.GetRoomResponse{
			RoomId:     r.ID,
			RoomNumber: r.RoomNumber,
			Floor:      int32(r.Floor),
			Capacity:   int32(r.Capacity),
			Occupied:   occupiedCount,
			Residents:  resInfos,
		}
	}
	
	return &roomv1.ListRoomsResponse{Rooms: respRooms}, nil
}

func (h *RoomHandler) AssignResident(ctx context.Context, req *roomv1.AssignResidentRequest) (*roomv1.AssignResidentResponse, error) {
	id, err := h.uc.AssignResident(ctx, req.UserId, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign resident: %v", err)
	}
	return &roomv1.AssignResidentResponse{ResidentId: id}, nil
}

func (h *RoomHandler) RemoveResident(ctx context.Context, req *roomv1.RemoveResidentRequest) (*roomv1.RemoveResidentResponse, error) {
	err := h.uc.RemoveResident(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove resident: %v", err)
	}
	return &roomv1.RemoveResidentResponse{Success: true}, nil
}

func (h *RoomHandler) GetResident(ctx context.Context, req *roomv1.GetResidentRequest) (*roomv1.GetResidentResponse, error) {
	r, err := h.uc.GetResident(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "resident not found: %v", err)
	}
	return &roomv1.GetResidentResponse{
		ResidentId:  r.ID,
		UserId:      r.UserID,
		RoomNumber:  r.RoomID,
		Role:        r.Role,
		CheckInDate: r.CheckInAt,
	}, nil
}

func (h *RoomHandler) ListResidents(ctx context.Context, req *roomv1.ListResidentsRequest) (*roomv1.ListResidentsResponse, error) {
	residents, err := h.uc.ListResidents(ctx, req.RoomId, req.Role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list residents: %v", err)
	}
	
	respRes := make([]*roomv1.GetResidentResponse, len(residents))
	for i, r := range residents {
		respRes[i] = &roomv1.GetResidentResponse{
			ResidentId:  r.ID,
			UserId:      r.UserID,
			RoomNumber:  r.RoomID,
			Role:        r.Role,
			CheckInDate: r.CheckInAt,
		}
	}
	return &roomv1.ListResidentsResponse{Residents: respRes, Total: int32(len(residents))}, nil
}

func (h *RoomHandler) UpdateResidentRole(ctx context.Context, req *roomv1.UpdateResidentRoleRequest) (*roomv1.UpdateResidentRoleResponse, error) {
	err := h.uc.UpdateResidentRole(ctx, req.UserId, req.NewRole)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}
	return &roomv1.UpdateResidentRoleResponse{Success: true}, nil
}

func (h *RoomHandler) GetDashboardStats(ctx context.Context, req *roomv1.GetDashboardStatsRequest) (*roomv1.GetDashboardStatsResponse, error) {
	stats, err := h.uc.GetDashboardStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}
	return &roomv1.GetDashboardStatsResponse{
		TotalResidents: int32(stats.TotalResidents),
		TotalRooms:     int32(stats.TotalRooms),
		OccupiedRooms:  int32(stats.OccupiedRooms),
		AvailableBeds:  int32(stats.AvailableBeds),
		OpenIssues:     0,
		UpcomingEvents: 0,
	}, nil
}
