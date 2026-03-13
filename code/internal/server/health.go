package server

import "net/http"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"status": "ok", "auth": s.authInfo()})
}

func (s *Server) handleAbilities(w http.ResponseWriter, r *http.Request) {
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"abilities": s.rg.ListCodes()})
}
