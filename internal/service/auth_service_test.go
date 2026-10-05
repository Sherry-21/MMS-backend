package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
	"maiora-backend-service/pkg/utils"
)

// mockUserRepository is a test mock for repository.UserRepository
type mockUserRepository struct {
	users map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	user, exists := m.users[username]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == userID {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

// mockRefreshTokenRepo is a test mock for repository.RefreshTokenRepository
type mockRefreshTokenRepo struct {
	tokens map[string]*model.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{
		tokens: make(map[string]*model.RefreshToken),
	}
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	m.tokens[token.RefreshToken] = token
	return nil
}

func (m *mockRefreshTokenRepo) FindByToken(ctx context.Context, tokenStr string) (*model.RefreshToken, error) {
	t, ok := m.tokens[tokenStr]
	if !ok {
		return nil, repository.ErrRefreshTokenNotFound
	}
	return t, nil
}

func (m *mockRefreshTokenRepo) RevokeToken(ctx context.Context, tokenStr string) error {
	t, ok := m.tokens[tokenStr]
	if !ok {
		return repository.ErrRefreshTokenNotFound
	}
	t.IsRevoked = true
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllUserTokens(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.IsRevoked = true
		}
	}
	return nil
}

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	authSvc := NewAuthService(mockRepo, mockTokens, "test-secret", 15*time.Minute, 7*24*time.Hour)

	req := &payloads.RegisterRequest{
		Username: "maiora_user",
		Password: "SecurePassword123!",
	}

	res, err := authSvc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Username != "maiora_user" {
		t.Errorf("expected username 'maiora_user', got %s", res.Username)
	}
	if res.UserID == "" {
		t.Errorf("expected user_id to be generated")
	}

	// Verify user is in repository with hashed password
	savedUser := mockRepo.users["maiora_user"]
	if savedUser == nil {
		t.Fatalf("expected user to be saved in repository")
	}
	if savedUser.PasswordHash == "SecurePassword123!" {
		t.Errorf("password was not hashed!")
	}
	if !utils.CheckPasswordHash("SecurePassword123!", savedUser.PasswordHash) {
		t.Errorf("password hash does not match original password")
	}
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	authSvc := NewAuthService(mockRepo, mockTokens, "test-secret", 15*time.Minute, 7*24*time.Hour)

	mockRepo.users["existing_user"] = &model.User{
		ID:        "user-1",
		Username:  "existing_user",
		CreatedAt: time.Now(),
	}

	req := &payloads.RegisterRequest{
		Username: "existing_user",
		Password: "Password123!",
	}

	_, err := authSvc.Register(context.Background(), req)
	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	authSvc := NewAuthService(mockRepo, mockTokens, "test-secret", 15*time.Minute, 7*24*time.Hour)

	hashedPassword, _ := utils.HashPassword("CorrectPassword123!")
	mockRepo.users["valid_user"] = &model.User{
		ID:           "usr-123",
		Username:     "valid_user",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	req := &payloads.LoginRequest{
		Username: "valid_user",
		Password: "CorrectPassword123!",
	}

	res, err := authSvc.Login(context.Background(), req)
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}

	if res.User.Username != "valid_user" {
		t.Errorf("expected username 'valid_user', got %s", res.User.Username)
	}
	if res.AccessToken == "" {
		t.Errorf("expected access_token to be populated")
	}
	if res.RefreshToken == "" {
		t.Errorf("expected refresh_token to be populated")
	}

	// Verify refresh token was saved in mockTokens table with is_revoked = false
	savedToken := mockTokens.tokens[res.RefreshToken]
	if savedToken == nil {
		t.Fatalf("refresh token was not saved to refresh_tokens table")
	}
	if savedToken.UserID != "usr-123" {
		t.Errorf("expected user_id 'usr-123', got %s", savedToken.UserID)
	}
	if savedToken.IsRevoked {
		t.Errorf("newly issued refresh token should not be revoked")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	authSvc := NewAuthService(mockRepo, mockTokens, "test-secret", 15*time.Minute, 7*24*time.Hour)

	hashedPassword, _ := utils.HashPassword("CorrectPassword123!")
	mockRepo.users["valid_user"] = &model.User{
		ID:           "usr-123",
		Username:     "valid_user",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	req := &payloads.LoginRequest{
		Username: "valid_user",
		Password: "WrongPassword!",
	}

	_, err := authSvc.Login(context.Background(), req)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	secret := "test-secret-refresh"
	authSvc := NewAuthService(mockRepo, mockTokens, secret, 15*time.Minute, 7*24*time.Hour)

	mockRepo.users["refresh_user"] = &model.User{
		ID:        "usr-999",
		Username:  "refresh_user",
		CreatedAt: time.Now(),
	}

	// Issue an initial token pair and save to mockTokens
	initialTokens, err := utils.GenerateTokenPair("usr-999", "refresh_user", secret, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate initial tokens: %v", err)
	}

	mockTokens.tokens[initialTokens.RefreshToken] = &model.RefreshToken{
		ID:           1,
		UserID:       "usr-999",
		RefreshToken: initialTokens.RefreshToken,
		ExpiresAt:    time.Now().UTC().Add(7 * 24 * time.Hour),
		IsRevoked:    false,
	}

	// Request refresh
	req := &payloads.RefreshTokenRequest{
		RefreshToken: initialTokens.RefreshToken,
	}

	newTokens, err := authSvc.RefreshToken(context.Background(), req)
	if err != nil {
		t.Fatalf("expected RefreshToken to succeed, got %v", err)
	}

	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Errorf("expected new access and refresh tokens")
	}
	if newTokens.RefreshToken == initialTokens.RefreshToken {
		t.Errorf("expected rotated refresh token to be different from original")
	}

	// Verify old refresh token is marked revoked
	oldTokenRecord := mockTokens.tokens[initialTokens.RefreshToken]
	if !oldTokenRecord.IsRevoked {
		t.Errorf("expected old refresh token to be marked is_revoked = true")
	}

	// Verify new token is stored and not revoked
	newTokenRecord := mockTokens.tokens[newTokens.RefreshToken]
	if newTokenRecord == nil {
		t.Fatalf("expected new refresh token to be saved in refresh_tokens table")
	}
	if newTokenRecord.IsRevoked {
		t.Errorf("new refresh token should not be revoked")
	}
}

func TestAuthService_RefreshToken_RevokedTokenReplay(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	secret := "test-secret-refresh"
	authSvc := NewAuthService(mockRepo, mockTokens, secret, 15*time.Minute, 7*24*time.Hour)

	mockRepo.users["refresh_user"] = &model.User{
		ID:       "usr-999",
		Username: "refresh_user",
	}

	tokens, err := utils.GenerateTokenPair("usr-999", "refresh_user", secret, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	// Token exists but is already revoked!
	mockTokens.tokens[tokens.RefreshToken] = &model.RefreshToken{
		ID:           1,
		UserID:       "usr-999",
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().UTC().Add(7 * 24 * time.Hour),
		IsRevoked:    true,
	}

	req := &payloads.RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	}

	_, err = authSvc.RefreshToken(context.Background(), req)
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken due to revoked token replay, got %v", err)
	}
}

func TestAuthService_RefreshToken_AccessTokenRejected(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockTokens := newMockRefreshTokenRepo()
	secret := "test-secret-refresh"
	authSvc := NewAuthService(mockRepo, mockTokens, secret, 15*time.Minute, 7*24*time.Hour)

	// Passing an access token instead of refresh token
	accessToken, err := utils.GenerateToken("usr-999", "refresh_user", utils.TokenTypeAccess, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	req := &payloads.RefreshTokenRequest{
		RefreshToken: accessToken,
	}

	_, err = authSvc.RefreshToken(context.Background(), req)
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken when providing access token, got %v", err)
	}
}
