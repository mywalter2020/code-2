package server

import (
	"net/http"
	"strconv"

	"juyu-ai-platform/internal/types"
)

func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := s.orc.ListTasks(types.TaskFilter{
		Status:   r.URL.Query().Get("status"),
		Scene:    r.URL.Query().Get("scene"),
		Operator: r.URL.Query().Get("operator"),
		Platform: r.URL.Query().Get("platform"),
		Query:    r.URL.Query().Get("q"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", types.PageResult{Items: items, Count: len(items), Limit: limit, Offset: offset})
}
