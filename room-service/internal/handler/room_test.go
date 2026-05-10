package handler

import (
	"context"
	"testing"
)

func TestUpdateResidentRoleValidation(t *testing.T) {
	h := &RoomHandler{} // Repo is nil

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
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
			err := h.UpdateResidentRole(context.Background(), "user1", tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateResidentRole() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err == nil || err.Error() != "rpc error: code = InvalidArgument desc = invalid role: "+tt.role {
				t.Errorf("Expected invalid argument error for %s, got: %v", tt.role, err)
			}
		})
	}
}
