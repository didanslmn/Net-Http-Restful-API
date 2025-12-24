package handler

import (
	"encoding/json"
	"net/http"
	"postgresDB/internal/delivery/middleware"
	"postgresDB/internal/delivery/response"
	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/service"
	"postgresDB/pkg/validator"

	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get user id from context
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		response.BadRequest(w, "ID user tidak valid")
		return
	}

	// get requester id and role from context
	requesterID, _ := middleware.GetUserID(r.Context())
	requesterRole, _ := middleware.GetUserRole(r.Context())

	// call service
	user, err := h.userService.GetUser(r.Context(), id, requesterID, requesterRole)
	if err != nil {
		// More specific error handling could be added here
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		//response.Error(w, err)
		return
	}
	response.Success(w, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get user id from path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// parse request body
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// validate request
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(w, err)
		return
	}

	// get requester id and role from context
	requesterID, _ := middleware.GetUserID(r.Context())
	requesterRole, _ := middleware.GetUserRole(r.Context())

	// call service
	updatedUser, err := h.userService.Update(r.Context(), id, requesterID, requesterRole, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, updatedUser)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get user id from path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	// parse request body
	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(w, err)
		return
	}

	// retrieve requester ID from context
	requesterID, _ := middleware.GetUserID(r.Context())

	// call service
	err = h.userService.ChangePassword(r.Context(), id, requesterID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, map[string]string{"message": "Password changed successfully"})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// retrieve requester ID from context
	requesterID, _ := middleware.GetUserID(r.Context())
	requesterRole, _ := middleware.GetUserRole(r.Context())

	// ensure requester is deleting their own account

	// call service
	err = h.userService.Delete(r.Context(), id, requesterID, requesterRole)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}
