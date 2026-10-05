package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"maiora-backend-service/internal/middleware"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/service"
	"maiora-backend-service/pkg/utils"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Registers a new user account with a unique username and securely hashed password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body payloads.RegisterRequest true "User Registration Details"
// @Success 201 {object} payloads.BaseResponse{data=payloads.UserResponse} "User created successfully"
// @Failure 400 {object} payloads.BaseResponse "Invalid request input or validation error"
// @Failure 409 {object} payloads.BaseResponse "Username is already taken"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req payloads.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	userResponse, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameRequired), errors.Is(err, service.ErrPasswordTooShort):
			utils.WriteBadRequest(w, err.Error(), nil)
		case errors.Is(err, service.ErrUsernameTaken):
			utils.WriteError(w, http.StatusConflict, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to register user", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, "User registered successfully", userResponse)
}

// Login godoc
// @Summary User login
// @Description Authenticates user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body payloads.LoginRequest true "User Login Credentials"
// @Success 200 {object} payloads.BaseResponse{data=payloads.LoginResponse} "Login successful"
// @Failure 400 {object} payloads.BaseResponse "Invalid request payload"
// @Failure 401 {object} payloads.BaseResponse "Invalid username or password"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req payloads.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	loginResponse, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid username or password", nil)
			return
		}
		utils.WriteInternalServerError(w, "Authentication error occurred", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Login successful", loginResponse)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchanges a valid refresh token for a newly rotated access token and refresh token pair
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body payloads.RefreshTokenRequest true "Refresh Token Payload"
// @Success 200 {object} payloads.BaseResponse{data=payloads.TokenResponse} "Token refreshed successfully"
// @Failure 400 {object} payloads.BaseResponse "Missing or invalid payload"
// @Failure 401 {object} payloads.BaseResponse "Invalid or expired refresh token"
// @Failure 500 {object} payloads.BaseResponse "Internal Server Error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req payloads.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request JSON payload", err.Error())
		return
	}

	tokenResponse, err := h.authService.RefreshToken(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRefreshTokenRequired):
			utils.WriteBadRequest(w, err.Error(), nil)
		case errors.Is(err, service.ErrInvalidRefreshToken):
			utils.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
		default:
			utils.WriteInternalServerError(w, "Failed to refresh token", err.Error())
		}
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Token refreshed successfully", tokenResponse)
}

// GetMe godoc
// @Summary Get current authenticated user profile
// @Description Retrieves the profile information of the currently authenticated user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} payloads.BaseResponse{data=payloads.UserResponse} "Authenticated user details"
// @Failure 401 {object} payloads.BaseResponse "Unauthorized"
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "User context not found", nil)
		return
	}

	response := user.ToUserResponse()
	utils.WriteSuccess(w, http.StatusOK, "User profile retrieved successfully", response)
}

