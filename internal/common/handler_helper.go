package common

import (
	"encoding/json"
	"net/http"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, model.ErrorResponse{Error: message})
}
