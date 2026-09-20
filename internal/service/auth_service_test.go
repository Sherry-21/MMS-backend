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

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	authSvc := NewAuthService(mockRepo)

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
	authSvc := NewAuthService(mockRepo)

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
	authSvc := NewAuthService(mockRepo)

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
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := newMockUserRepository()
	authSvc := NewAuthService(mockRepo)

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
