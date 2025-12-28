package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"postgresDB/internal/delivery/middleware"
	"postgresDB/internal/delivery/response"
	"postgresDB/internal/domain/dto"
	"postgresDB/internal/domain/service"

	"postgresDB/pkg/validator"

	"github.com/google/uuid"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// method check
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format JSON tidak valid")
		return
	}

	// validation struct
	if err := validator.ValidateStruct(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	//
	product, err := h.productService.Create(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Created(w, product)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// check mthod
	if r.Method != http.MethodGet {
		response.BadRequest(w, "Method not allowed")
	}

	// Extract ID from url path
	idStr := r.PathValue("id")
	if idStr == "" {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}

	//
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "ID produk tidak valid")
	}

	product, err := h.productService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, product)
}

// List handles listing products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodGet {
		response.BadRequest(w, "Method not allowed")
		return
	}

	req := dto.ProductListRequest{
		Page:     parseIntQuery(r, "page", 1),
		Limit:    parseIntQuery(r, "limit", 10),
		Search:   r.URL.Query().Get("search"),
		Category: r.URL.Query().Get("category"),
	}

	products, meta, err := h.productService.List(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.SuccessWithMeta(w, products, meta)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// extract id from url path
	idStr := r.PathValue("id")
	if idStr == "" {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}
	userRole, err := middleware.GetUserRole(r.Context())
	if err != nil {
		response.BadRequest(w, "Role tidak ditemukan")
		return
	}
	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Format JSON tidak valid")
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	product, err := h.productService.Update(r.Context(), id, &req, userRole)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, product)

}

// Delete handles deleting a product
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Extract ID from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}

	userRole, err := middleware.GetUserRole(r.Context())
	if err != nil {
		response.BadRequest(w, "Role tidak ditemukan")
		return
	}

	if err := h.productService.Delete(r.Context(), id, userRole); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}
