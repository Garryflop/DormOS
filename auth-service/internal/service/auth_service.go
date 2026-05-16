package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"golang.org/x/crypto/bcrypt"

	"github.com/Garryflop/DormManage/auth-service/internal/cache"
	"github.com/Garryflop/DormManage/auth-service/internal/config"
	"github.com/Garryflop/DormManage/auth-service/internal/domain"
	"github.com/Garryflop/DormManage/auth-service/internal/repository"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	cache       *cache.RedisCache
	cfg         *config.Config
	nats        *nats.Conn
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	cache *cache.RedisCache,
	cfg *config.Config,
	natsConn *nats.Conn,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cache:       cache,
		cfg:         cfg,
		nats:        natsConn,
	}
}

type RegisterResult struct {
	User *domain.User
}

func (s *AuthService) Register(ctx context.Context, input domain.RegisterInput) (*RegisterResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		Role:         input.Role,
		DormID:       input.DormID,
		RoomNumber:   input.RoomNumber,
		Floor:        input.Floor,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	_ = s.publishUserCreated(created)

	return &RegisterResult{User: created}, nil
}

type userCreatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	DormID string `json:"dorm_id"`
}

func (s *AuthService) publishUserCreated(user *domain.User) error {
	event := userCreatedEvent{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
		DormID: user.DormID.String(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.nats.Publish("auth.user.created", data)
}

type LoginResult struct {
	AccessToken       string
	RefreshToken      string
	AccessTokenExpiry time.Time
	User              *domain.User
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidPassword
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, domain.ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	accessExpiry := time.Now().Add(s.cfg.AccessTokenDuration)
	accessToken, err := s.generateAccessToken(user, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(s.cfg.RefreshTokenDuration),
		CreatedAt:    now,
		Revoked:      false,
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	claims := &cache.CachedTokenClaims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		DormID: user.DormID.String(),
		Email:  user.Email,
	}
	ttl := time.Until(accessExpiry)
	_ = s.cache.SetTokenValidation(ctx, accessToken, claims, ttl)

	return &LoginResult{
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		AccessTokenExpiry: accessExpiry,
		User:              user,
	}, nil
}

type TokenClaims struct {
	UserID string
	Role   string
	DormID string
	Email  string
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	cached, err := s.cache.GetTokenValidation(ctx, tokenString)
	if err == nil && cached != nil {
		return &TokenClaims{
			UserID: cached.UserID,
			Role:   cached.Role,
			DormID: cached.DormID,
			Email:  cached.Email,
		}, nil
	}

	claims, err := s.parseAccessToken(tokenString)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	cachedClaims := &cache.CachedTokenClaims{
		UserID: claims.UserID,
		Role:   claims.Role,
		DormID: claims.DormID,
		Email:  claims.Email,
	}
	expiry, _ := claims.raw.GetExpirationTime()
	if expiry != nil {
		ttl := time.Until(expiry.Time)
		if ttl > 0 {
			_ = s.cache.SetTokenValidation(ctx, tokenString, cachedClaims, ttl)
		}
	}

	return &TokenClaims{
		UserID: claims.UserID,
		Role:   claims.Role,
		DormID: claims.DormID,
		Email:  claims.Email,
	}, nil
}

type RefreshResult struct {
	AccessToken       string
	RefreshToken      string
	AccessTokenExpiry time.Time
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if !session.IsValid() {
		if session.Revoked {
			return nil, domain.ErrSessionRevoked
		}
		return nil, domain.ErrSessionExpired
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, domain.ErrAccountInactive
	}

	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("revoke old session: %w", err)
	}

	accessExpiry := time.Now().Add(s.cfg.AccessTokenDuration)
	accessToken, err := s.generateAccessToken(user, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	newSession := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    now.Add(s.cfg.RefreshTokenDuration),
		CreatedAt:    now,
	}
	if _, err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("save new session: %w", err)
	}

	return &RefreshResult{
		AccessToken:       accessToken,
		RefreshToken:      newRefreshToken,
		AccessTokenExpiry: accessExpiry,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return domain.ErrSessionNotFound
	}
	return s.sessionRepo.Revoke(ctx, session.ID)
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

type jwtClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	DormID string `json:"dorm"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type parsedClaims struct {
	UserID string
	Role   string
	DormID string
	Email  string
	raw    jwt.Claims
}

func (s *AuthService) generateAccessToken(user *domain.User, expiry time.Time) (string, error) {
	claims := jwtClaims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		DormID: user.DormID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) parseAccessToken(tokenString string) (*parsedClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	return &parsedClaims{
		UserID: c.UserID,
		Role:   c.Role,
		DormID: c.DormID,
		Email:  c.Email,
		raw:    c,
	}, nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
