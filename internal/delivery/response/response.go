package response

import (
	"encoding/json"
	"net/http"

	"postgresDB/internal/domain/dto"
	apperrors "postgresDB/internal/domain/errors"
)

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// Success writes a success response
func Success(w http.ResponseWriter, data interface{}) {
	resp := dto.NewSuccessResponse(data)
	JSON(w, http.StatusOK, resp)
}

// SuccessWithMeta writes a success response with metadata
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta interface{}) {
	resp := dto.NewSuccessResponseWithMeta(data, meta)
	JSON(w, http.StatusOK, resp)
}

// Created writes a created response
func Created(w http.ResponseWriter, data interface{}) {
	resp := dto.NewSuccessResponse(data)
	JSON(w, http.StatusCreated, resp)
}

// NoContent writes a no content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error response
func Error(w http.ResponseWriter, err error) {
	appErr, ok := apperrors.AsAppError(err)
	if !ok {
		appErr = apperrors.ErrInternal
	}

	var details []dto.ValidationError
	if len(appErr.Details) > 0 {
		details = make([]dto.ValidationError, len(appErr.Details))
		for i, d := range appErr.Details {
			details[i] = dto.ValidationError{
				Field:   d.Field,
				Message: d.Message,
			}
		}
	}

	resp := dto.NewErrorResponse(string(appErr.Code), appErr.Message, details)
	JSON(w, appErr.HTTPStatus, resp)
}

// BadRequest writes a bad request error
func BadRequest(w http.ResponseWriter, message string) {
	resp := dto.NewErrorResponse(string(apperrors.CodeBadRequest), message, nil)
	JSON(w, http.StatusBadRequest, resp)
}
