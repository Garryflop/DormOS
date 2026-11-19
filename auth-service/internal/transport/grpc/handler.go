package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
	authv1 "github.com/Garryflop/DormOS-gen-go/auth/v1"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	uc domain.AuthUseCase
}

func NewAuthHandler(uc domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	user, err := h.uc.Register(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RegisterResponse{
		UserId: user.ID.String(),
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	accessToken, refreshToken, err := h.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, role, err := h.uc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}
	return &authv1.ValidateTokenResponse{
		IsValid: true,
		UserId:  userID,
		Role:    role,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	newAccess, newRefresh, err := h.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RefreshTokenResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.uc.Logout(ctx, req.AccessToken); err != nil {
		return nil, mapError(err)
	}
	return &authv1.LogoutResponse{Success: true}, nil
}

func (h *AuthHandler) GetProfile(ctx context.Context, req *authv1.GetProfileRequest) (*authv1.GetProfileResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}

	user, err := h.uc.GetProfile(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.GetProfileResponse{
		UserId:      user.ID.String(),
		Email:       user.Email,
		FullName:    user.FullName,
		Phone:       user.Phone,
		Role:        string(user.Role),
		AvatarUrl:   user.AvatarURL,
		IsSuspended: user.IsSuspended,
		CreatedAt:   user.CreatedAt.Unix(),
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Errorf(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Errorf(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Errorf(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrAccountSuspended):
		return status.Errorf(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Errorf(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrSessionRevoked),
		errors.Is(err, domain.ErrSessionExpired):
		return status.Errorf(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrEmailRequired),
		errors.Is(err, domain.ErrPasswordRequired),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrFullNameRequired):
		return status.Errorf(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
