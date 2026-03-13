package server

import (
	"net/http"
	"strings"
)

func (s *Server) requireWriteAuth(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(s.apiKey) == "" {
		return true
	}
	provided := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if provided == "" {
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			provided = strings.TrimSpace(authz[7:])
		}
	}
	if provided == s.apiKey {
		return true
	}
	writeAPI(w, http.StatusUnauthorized, false, statusCodeToErr(http.StatusUnauthorized), "", "missing or invalid api key", map[string]any{
		"hint":    "set X-API-Key or Authorization: Bearer <key>",
		"enabled": true,
	})
	return false
}

func (s *Server) authInfo() map[string]any {
	return map[string]any{
		"api_key_enabled": strings.TrimSpace(s.apiKey) != "",
		"write_header":    "X-API-Key",
	}
}
