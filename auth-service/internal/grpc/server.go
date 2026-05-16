package grpc

import (
	"context"

	authv1 "github.com/Garryflop/DormOS-gen-go/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer is a dummy implementation to silence compilation errors.
// Diasbek needs to implement the actual logic here.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
}

func NewAuthServer() *AuthServer {
	return &AuthServer{}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *AuthServer) ResetPasswordRequest(ctx context.Context, req *authv1.ResetPasswordRequestRequest) (*authv1.ResetPasswordRequestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetPasswordRequest not implemented")
}

func (s *AuthServer) ResetPasswordConfirm(ctx context.Context, req *authv1.ResetPasswordConfirmRequest) (*authv1.ResetPasswordConfirmResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetPasswordConfirm not implemented")
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateToken not implemented")
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshToken not implemented")
}

func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}

func (s *AuthServer) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}

func (s *AuthServer) GetProfile(ctx context.Context, req *authv1.GetProfileRequest) (*authv1.GetProfileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProfile not implemented")
}

func (s *AuthServer) UpdateProfile(ctx context.Context, req *authv1.UpdateProfileRequest) (*authv1.UpdateProfileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProfile not implemented")
}

func (s *AuthServer) SuspendAccount(ctx context.Context, req *authv1.SuspendAccountRequest) (*authv1.SuspendAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SuspendAccount not implemented")
}

func (s *AuthServer) ActivateAccount(ctx context.Context, req *authv1.ActivateAccountRequest) (*authv1.ActivateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ActivateAccount not implemented")
}
