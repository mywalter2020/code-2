package server

import (
	"fmt"
	"net/http"
	"strings"
)

type operatorIdentity struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type operatorAuth struct {
	required bool
	tokens   map[string]string
}

func newOperatorAuth(raw string) operatorAuth {
	items := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pieces := strings.SplitN(part, ":", 2)
		if len(pieces) != 2 {
			continue
		}
		id := strings.TrimSpace(pieces[0])
		token := strings.TrimSpace(pieces[1])
		if id == "" || token == "" {
			continue
		}
		items[id] = token
	}
	return operatorAuth{required: len(items) > 0, tokens: items}
}

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

func (s *Server) requireOperator(w http.ResponseWriter, r *http.Request) (operatorIdentity, bool) {
	identity := operatorIdentity{
		ID:   strings.TrimSpace(r.Header.Get("X-Operator-ID")),
		Name: strings.TrimSpace(r.Header.Get("X-Operator-Name")),
	}
	if !s.operatorAuth.required {
		return identity, true
	}
	if identity.ID == "" {
		writeAPI(w, http.StatusUnauthorized, false, statusCodeToErr(http.StatusUnauthorized), "", "missing operator identity", map[string]any{
			"hint":           "set X-Operator-ID and X-Operator-Token",
			"operator_auth":  true,
			"identity_header": "X-Operator-ID",
		})
		return operatorIdentity{}, false
	}
	provided := strings.TrimSpace(r.Header.Get("X-Operator-Token"))
	expected := s.operatorAuth.tokens[identity.ID]
	if expected == "" || provided == "" || provided != expected {
		writeAPI(w, http.StatusUnauthorized, false, statusCodeToErr(http.StatusUnauthorized), "", "missing or invalid operator token", map[string]any{
			"hint":            "set X-Operator-ID and X-Operator-Token",
			"operator_auth":   true,
			"operator_id":     identity.ID,
			"known_operator":  expected != "",
			"identity_header": "X-Operator-ID",
		})
		return operatorIdentity{}, false
	}
	return identity, true
}

func pickOperator(explicit string, identity operatorIdentity) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit == "" {
		return identity.ID, nil
	}
	if identity.ID != "" && explicit != identity.ID {
		return "", fmt.Errorf("operator mismatch: request=%s header=%s", explicit, identity.ID)
	}
	return explicit, nil
}

func (s *Server) authInfo() map[string]any {
	return map[string]any{
		"api_key_enabled":        strings.TrimSpace(s.apiKey) != "",
		"write_header":           "X-API-Key",
		"operator_auth_enabled":  s.operatorAuth.required,
		"operator_identity_header": "X-Operator-ID",
		"operator_token_header":  "X-Operator-Token",
	}
}
