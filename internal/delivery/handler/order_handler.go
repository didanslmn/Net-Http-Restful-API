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

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// method check
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.BadRequest(w, "User tidak ditemukan")
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	order, err := h.orderService.Create(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, order)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	// method check
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.BadRequest(w, "User tidak ditemukan")
		return
	}

	UserRole, err := middleware.GetUserRole(r.Context())
	if err != nil {
		response.BadRequest(w, "Role tidak ditemukan")
		return
	}
	// Extract id from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		response.BadRequest(w, "ID tidak ditemukan")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "ID tidak valid")
		return
	}

	order, err := h.orderService.GetByID(r.Context(), id, UserID, UserRole)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// method check
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.BadRequest(w, "User tidak ditemukan")
		return
	}

	userRole, err := middleware.GetUserRole(r.Context())
	if err != nil {
		response.BadRequest(w, "Role tidak ditemukan")
		return
	}

	req := dto.OrderListRequest{
		Page:   parseIntQuery(r, "page", 1),
		Limit:  parseIntQuery(r, "limit", 10),
		Status: r.URL.Query().Get("status"),
	}

	if err := validator.ValidateStruct(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	orders, pagination, err := h.orderService.ListAll(r.Context(), userID, userRole, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.SuccessWithMeta(w, orders, pagination)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		response.BadRequest(w, "Method not allowed")
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.BadRequest(w, "User tidak ditemukan")
		return
	}

	userRole, err := middleware.GetUserRole(r.Context())
	if err != nil {
		response.BadRequest(w, "Role tidak ditemukan")
		return
	}

	// Extract ID from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		response.BadRequest(w, "ID order tidak valid")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "ID order tidak valid")
		return
	}

	var req dto.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format JSON tidak valid")
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(w, err)
		return
	}

	order, err := h.orderService.UpdateStatus(r.Context(), id, userID, userRole, &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, order)
}

// parseOrderIntQuery parses an integer query parameter with a default value
// func parseOrderIntQuery(r *http.Request, key string, defaultValue int) int {
// 	val := r.URL.Query().Get(key)
// 	if val == "" {
// 		return defaultValue
// 	}
// 	intVal, err := strconv.Atoi(val)
// 	if err != nil {
// 		return defaultValue
// 	}
// 	return intVal
// }
