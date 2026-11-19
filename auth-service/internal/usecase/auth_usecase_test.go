package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
	"github.com/Garryflop/DormManage/auth-service/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ─── Mock Repositories ────────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	args := m.Called(ctx, session)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockSessionRepo) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockSessionRepo) GetByAccessToken(ctx context.Context, token string) (*domain.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockSessionRepo) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	return m.Called(ctx, sessionID).Error(0)
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

const testJwtSecret = "super-secret-key-12345-super-secret-key"

// ─── Register Tests ──────────────────────────────────────────────────────────

func TestAuthUseCase_Register_Success(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	uRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*domain.User)
			assert.Equal(t, "test@test.com", user.Email)
			assert.Equal(t, "John Doe", user.FullName)
			assert.Equal(t, domain.RoleStudent, user.Role)
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
			assert.NoError(t, err)
		}).
		Return(&domain.User{
			ID:       uuid.New(),
			Email:    "test@test.com",
			FullName: "John Doe",
			Role:     domain.RoleStudent,
		}, nil)

	res, err := uc.Register(context.Background(), "test@test.com", "password123", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "test@test.com", res.Email)
	uRepo.AssertExpectations(t)
}

func TestAuthUseCase_Register_ValidationErrors(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	// Empty email
	_, err := uc.Register(context.Background(), "", "password123", "John Doe")
	assert.ErrorIs(t, err, domain.ErrEmailRequired)

	// Empty password
	_, err = uc.Register(context.Background(), "test@test.com", "", "John Doe")
	assert.ErrorIs(t, err, domain.ErrPasswordRequired)

	// Password too short
	_, err = uc.Register(context.Background(), "test@test.com", "short", "John Doe")
	assert.ErrorIs(t, err, domain.ErrPasswordTooShort)

	// Empty full name
	_, err = uc.Register(context.Background(), "test@test.com", "password123", "")
	assert.ErrorIs(t, err, domain.ErrFullNameRequired)
}

// ─── Login Tests ─────────────────────────────────────────────────────────────

func TestAuthUseCase_Login_Success(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	uID := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	user := &domain.User{
		ID:       uID,
		Email:    "test@test.com",
		Password: string(hash),
		Role:     domain.RoleStudent,
	}

	uRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(user, nil)
	sRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).
		Return(&domain.Session{}, nil)

	accessToken, refreshToken, err := uc.Login(context.Background(), "test@test.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAuthUseCase_Login_Suspended(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	user := &domain.User{
		Email:       "test@test.com",
		IsSuspended: true,
	}

	uRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(user, nil)

	_, _, err := uc.Login(context.Background(), "test@test.com", "password123")
	assert.ErrorIs(t, err, domain.ErrAccountSuspended)
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	user := &domain.User{
		Email:    "test@test.com",
		Password: string(hash),
	}

	uRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(user, nil)

	_, _, err := uc.Login(context.Background(), "test@test.com", "wrongpassword")
	assert.ErrorIs(t, err, domain.ErrInvalidPassword)
}

// ─── ValidateToken Tests ─────────────────────────────────────────────────────

func TestAuthUseCase_ValidateToken_SuccessAndFail(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	uID := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	user := &domain.User{
		ID:       uID,
		Email:    "test@test.com",
		Password: string(hash),
		Role:     domain.RoleStudent,
	}

	uRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(user, nil)
	sRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).
		Return(&domain.Session{}, nil)

	accessToken, _, err := uc.Login(context.Background(), "test@test.com", "password123")
	assert.NoError(t, err)

	// Validate valid token
	validatedID, role, err := uc.ValidateToken(context.Background(), accessToken)
	assert.NoError(t, err)
	assert.Equal(t, uID.String(), validatedID)
	assert.Equal(t, string(domain.RoleStudent), role)

	// Validate invalid token
	_, _, err = uc.ValidateToken(context.Background(), "invalid-token-string")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

// ─── RefreshToken Tests ──────────────────────────────────────────────────────

func TestAuthUseCase_RefreshToken_Success(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	uID := uuid.New()
	sID := uuid.New()
	session := &domain.Session{
		ID:           sID,
		UserID:       uID,
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Revoked:      false,
	}
	user := &domain.User{
		ID:    uID,
		Email: "test@test.com",
		Role:  domain.RoleStudent,
	}

	sRepo.On("GetByRefreshToken", mock.Anything, "old-refresh").Return(session, nil)
	uRepo.On("GetByID", mock.Anything, uID).Return(user, nil)
	sRepo.On("Revoke", mock.Anything, sID).Return(nil)
	sRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(&domain.Session{}, nil)

	newAccess, newRefresh, err := uc.RefreshToken(context.Background(), "old-refresh")
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
}

func TestAuthUseCase_RefreshToken_SessionExpired(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	session := &domain.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		Revoked:   false,
	}

	sRepo.On("GetByRefreshToken", mock.Anything, "expired-refresh").Return(session, nil)

	_, _, err := uc.RefreshToken(context.Background(), "expired-refresh")
	assert.ErrorIs(t, err, domain.ErrSessionExpired)
}

func TestAuthUseCase_RefreshToken_SessionRevoked(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	session := &domain.Session{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Revoked:   true, // Revoked
	}

	sRepo.On("GetByRefreshToken", mock.Anything, "revoked-refresh").Return(session, nil)

	_, _, err := uc.RefreshToken(context.Background(), "revoked-refresh")
	assert.ErrorIs(t, err, domain.ErrSessionRevoked)
}

// ─── Logout Tests ────────────────────────────────────────────────────────────

func TestAuthUseCase_Logout_Success(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	sID := uuid.New()
	sRepo.On("GetByAccessToken", mock.Anything, "token-123").Return(&domain.Session{ID: sID}, nil)
	sRepo.On("Revoke", mock.Anything, sID).Return(nil)

	err := uc.Logout(context.Background(), "token-123")
	assert.NoError(t, err)
}

func TestAuthUseCase_Logout_NotFound(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	sRepo.On("GetByAccessToken", mock.Anything, "token-123").Return(nil, errors.New("not found"))

	err := uc.Logout(context.Background(), "token-123")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

// ─── GetProfile Tests ────────────────────────────────────────────────────────

func TestAuthUseCase_GetProfile_Success(t *testing.T) {
	uRepo := &mockUserRepo{}
	sRepo := &mockSessionRepo{}
	uc := usecase.NewAuthUseCase(uRepo, sRepo, testJwtSecret)

	uID := uuid.New()
	user := &domain.User{ID: uID, Email: "test@test.com"}
	uRepo.On("GetByID", mock.Anything, uID).Return(user, nil)

	res, err := uc.GetProfile(context.Background(), uID)
	assert.NoError(t, err)
	assert.Equal(t, uID, res.ID)
}
