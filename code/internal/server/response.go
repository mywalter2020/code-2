package server

import (
	"encoding/json"
	"net/http"

	"juyu-ai-platform/internal/types"
)

func writeAPI(w http.ResponseWriter, status int, success bool, code, message, errMsg string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.APIResponse{
		Success: success,
		Code:    code,
		Message: message,
		Error:   errMsg,
		Data:    data,
	})
}
