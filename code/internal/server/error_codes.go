package server

import "juyu-ai-platform/internal/types"

func statusCodeToErr(status int) string {
	switch status {
	case 400:
		return types.ErrCodeBadRequest
	case 404:
		return types.ErrCodeNotFound
	case 405:
		return types.ErrCodeMethodNotAllowed
	default:
		return types.ErrCodeInternal
	}
}
