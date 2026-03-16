package adapters

import "net/http"

func httpMethodForAction(action string) string {
	switch action {
	case "publish", "on_shelf":
		return http.MethodPost
	case "update":
		return http.MethodPut
	case "off_shelf":
		return http.MethodPost
	default:
		return http.MethodPost
	}
}
