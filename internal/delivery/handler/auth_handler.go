package handler

import (
	"encoding/json"
	"net/http"
	"postgresDB/internal/delivery/middleware"
	"postgresDB/internal/delivery/response"
	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/service"
	"postgresDB/pkg/validator"
	"time"
)

type AuthHandler struct {
	authService service.AuthService
	refreshTTL  time.Duration
}

func NewAuthHandler(authService service.AuthService, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		refreshTTL:  refreshTTL,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// method POST check
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse request body
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		//response.BadRequest(w,"Invalit rekuest body")
		return
	}

	// validate request
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(w, err)
		return
	}

	// call service
	res, err := h.authService.Register(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	// set refresh token in http-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(h.refreshTTL),
		HttpOnly: true,
		Path:     "/",
	})

	response.Created(w, res)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// method POST check
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// parse request body
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(w, err)
		return
	}
	// call service
	res, err := h.authService.Login(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}
	// set refresh token in http-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(h.refreshTTL),
		HttpOnly: true,
		Path:     "/",
	})

	response.Success(w, res)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// method POST check
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get access token JTI and exp from context (set by auth middleware)
	jti, _ := middleware.GetTokenJTIFromContext(r.Context())

	exp, ok := r.Context().Value(middleware.TokenExpKey).(time.Time)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value

	// Call service to blacklist the token
	if err := h.authService.Logout(r.Context(), jti, exp, refreshToken); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	response.Success(w, map[string]string{"message": "logout berhasil"})
}

// RefreshToken handles token refresh

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// method POST check
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}
	// get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value
	// call service
	res, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		// clear refresh token cookie on error
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Path:     "/",
		})
		response.Error(w, err)
		return
	}

	// set new refresh token in http-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(h.refreshTTL),
		HttpOnly: true,
		Path:     "/",
	})

	response.Success(w, res)
}

// RevokeAllSessions handles revoking all user sessions
func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	// method POST check
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.BadRequest(w, "user tidak ditemukan")
		return
	}

	// Call service to revoke all sessions
	if err := h.authService.RevokeAllSessions(r.Context(), userID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response.Success(w, map[string]string{"message": "All sessions revoked successfully"})
}
