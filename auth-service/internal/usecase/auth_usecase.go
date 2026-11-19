package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
)

type authUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	jwtSecret string,
) domain.AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
		accessTTL:   15 * time.Minute,
		refreshTTL:  7 * 24 * time.Hour,
	}
}

func (uc *authUseCase) Register(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	if email == "" {
		return nil, domain.ErrEmailRequired
	}
	if password == "" {
		return nil, domain.ErrPasswordRequired
	}
	if len(password) < 8 {
		return nil, domain.ErrPasswordTooShort
	}
	if fullName == "" {
		return nil, domain.ErrFullNameRequired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hash),
		FullName:  fullName,
		Role:      domain.RoleStudent,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return uc.userRepo.Create(ctx, user)
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", domain.ErrInvalidPassword
	}

	if user.IsSuspended {
		return "", "", domain.ErrAccountSuspended
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", domain.ErrInvalidPassword
	}

	accessToken, err := uc.generateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateSecureToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(uc.refreshTTL),
		CreatedAt:    now,
	}
	if _, err := uc.sessionRepo.Create(ctx, session); err != nil {
		return "", "", fmt.Errorf("save session: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (uc *authUseCase) ValidateToken(ctx context.Context, accessToken string) (string, string, error) {
	claims, err := uc.parseToken(accessToken)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}
	return claims.UserID, claims.Role, nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := uc.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", domain.ErrSessionNotFound
	}

	if !session.IsValid() {
		if session.Revoked {
			return "", "", domain.ErrSessionRevoked
		}
		return "", "", domain.ErrSessionExpired
	}

	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", err
	}

	if err := uc.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return "", "", fmt.Errorf("revoke session: %w", err)
	}

	newAccess, err := uc.generateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRefresh, err := generateSecureToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	newSession := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newRefresh,
		ExpiresAt:    now.Add(uc.refreshTTL),
		CreatedAt:    now,
	}
	if _, err := uc.sessionRepo.Create(ctx, newSession); err != nil {
		return "", "", fmt.Errorf("save session: %w", err)
	}

	return newAccess, newRefresh, nil
}

func (uc *authUseCase) Logout(ctx context.Context, accessToken string) error {
	session, err := uc.sessionRepo.GetByAccessToken(ctx, accessToken)
	if err != nil {
		return domain.ErrSessionNotFound
	}
	return uc.sessionRepo.Revoke(ctx, session.ID)
}

func (uc *authUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

type jwtClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (uc *authUseCase) generateAccessToken(user *domain.User) (string, error) {
	claims := jwtClaims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *authUseCase) parseToken(tokenString string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}
	c, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	return c, nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
