package grpc

import (
	"context"
	"testing"
	
	"github.com/Garryflop/DormManage/room-service/internal/domain"
	roomv1 "github.com/Garryflop/DormOS-gen-go/room/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockUseCase is a simple mock to test the transport layer
type mockUseCase struct {
	domain.RoomUseCase
}

func (m *mockUseCase) UpdateResidentRole(ctx context.Context, userID, role string) error {
	validRoles := map[string]bool{"student": true, "manager": true, "admin": true}
	if !validRoles[role] {
		return status.Errorf(codes.InvalidArgument, "invalid role: %s", role)
	}
	return nil
}

func TestUpdateResidentRoleValidation(t *testing.T) {
	uc := &mockUseCase{}
	h := NewRoomHandler(uc)

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name:    "valid role student",
			role:    "student",
			wantErr: false,
		},
		{
			name:    "invalid role superadmin",
			role:    "superadmin",
			wantErr: true,
		},
		{
			name:    "invalid role empty string",
			role:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := h.UpdateResidentRole(context.Background(), &roomv1.UpdateResidentRoleRequest{
				UserId:  "user1",
				NewRole: tt.role,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateResidentRole() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.wantErr {
				expectedErr := "rpc error: code = Internal desc = failed to update role: rpc error: code = InvalidArgument desc = invalid role: " + tt.role
				if err == nil || err.Error() != expectedErr {
					t.Errorf("Expected invalid argument error for %s, got: %v", tt.role, err)
				}
			}
		})
	}
}
