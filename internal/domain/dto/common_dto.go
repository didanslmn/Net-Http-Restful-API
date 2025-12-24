package dto

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information in API response
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents a validation error detail
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// NewSuccessResponseWithMeta creates a success response with metadata
func NewSuccessResponseWithMeta(data interface{}, meta interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string, details []ValidationError) APIResponse {
	return APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// CalculateTotalPages calculates total pages from total items and limit
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}
	return pages
}
