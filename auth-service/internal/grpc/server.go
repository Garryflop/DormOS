package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
	"github.com/Garryflop/DormManage/auth-service/internal/service"
	authv1 "github.com/Garryflop/DormOS-gen-go/auth/v1"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func NewServer(svc *service.AuthService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	dormID, err := uuid.Parse(req.DormId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dorm_id: %v", err)
	}

	input := domain.RegisterInput{
		Email:      req.Email,
		Password:   req.Password,
		FullName:   req.FullName,
		Role:       domain.Role(req.Role),
		DormID:     dormID,
		RoomNumber: req.RoomNumber,
		Floor:      int(req.Floor),
	}

	result, err := s.svc.Register(ctx, input)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RegisterResponse{
		UserId:    result.User.ID.String(),
		Email:     result.User.Email,
		Role:      string(result.User.Role),
		CreatedAt: result.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	result, err := s.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.LoginResponse{
		AccessToken:       result.AccessToken,
		RefreshToken:      result.RefreshToken,
		AccessTokenExpiry: result.AccessTokenExpiry.Unix(),
		User:              toProtoProfile(result.User),
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := s.svc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Role:   claims.Role,
		DormId: claims.DormID,
		Email:  claims.Email,
	}, nil
}

func (s *Server) GetUserProfile(ctx context.Context, req *authv1.GetUserProfileRequest) (*authv1.UserProfileResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user, err := s.svc.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.UserProfileResponse{Profile: toProtoProfile(user)}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	result, err := s.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:       result.AccessToken,
		RefreshToken:      result.RefreshToken,
		AccessTokenExpiry: result.AccessTokenExpiry.Unix(),
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.svc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, mapError(err)
	}
	return &authv1.LogoutResponse{Success: true}, nil
}

func toProtoProfile(u *domain.User) *authv1.UserProfile {
	return &authv1.UserProfile{
		UserId:     u.ID.String(),
		Email:      u.Email,
		FullName:   u.FullName,
		Role:       string(u.Role),
		DormId:     u.DormID.String(),
		RoomNumber: u.RoomNumber,
		Floor:      int32(u.Floor),
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Errorf(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Errorf(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Errorf(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrAccountInactive):
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
		errors.Is(err, domain.ErrFullNameRequired),
		errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrDormIDRequired):
		return status.Errorf(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
