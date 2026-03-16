package adapters

import (
	"fmt"
	"strings"

	"juyu-ai-platform/internal/types"
)

type AlibabaItemPayload struct {
	OutBizID    string            `json:"out_biz_id,omitempty"`
	Title       string            `json:"title"`
	Subtitle    string            `json:"subtitle,omitempty"`
	Description string            `json:"description,omitempty"`
	Category    string            `json:"category,omitempty"`
	Brand       string            `json:"brand,omitempty"`
	SKU         string            `json:"sku,omitempty"`
	Price       float64           `json:"price,omitempty"`
	Currency    string            `json:"currency,omitempty"`
	Images      []string          `json:"images,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Content     string            `json:"content,omitempty"`
	Page        map[string]any    `json:"page,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	PublishMode string            `json:"publish_mode,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func alibabaItemPayload(req types.AdapterRequest) (AlibabaItemPayload, error) {
	product, err := extractProduct(req.Payload)
	if err != nil {
		return AlibabaItemPayload{}, err
	}
	publish := extractPublish(req.Payload)
	metadata := map[string]string{}
	for k, v := range publish.Metadata {
		metadata[k] = fmt.Sprint(v)
	}
	if req.ExternalRef != "" {
		metadata["external_ref"] = req.ExternalRef
	}
	return AlibabaItemPayload{
		OutBizID:    firstNonEmpty(req.ExternalRef, product.SKU, product.ID),
		Title:       product.Title,
		Subtitle:    product.Subtitle,
		Description: product.Description,
		Category:    product.Category,
		Brand:       product.Brand,
		SKU:         product.SKU,
		Price:       product.Price,
		Currency:    product.Currency,
		Images:      product.Images,
		Tags:        product.Tags,
		Content:     publish.Content,
		Page:        publish.Page,
		Channel:     publish.Channel,
		PublishMode: publish.PublishMode,
		Metadata:    metadata,
	}, nil
}

func extractProduct(payload map[string]any) (types.Product, error) {
	if payload == nil {
		return types.Product{}, fmt.Errorf("missing payload")
	}
	switch v := payload["product"].(type) {
	case types.Product:
		if strings.TrimSpace(v.Title) == "" {
			return types.Product{}, fmt.Errorf("missing product.title")
		}
		return v, nil
	case map[string]any:
		p := types.Product{
			ID:          anyString(v["id"]),
			SKU:         anyString(v["sku"]),
			Title:       anyString(v["title"]),
			Subtitle:    anyString(v["subtitle"]),
			Description: anyString(v["description"]),
			Category:    anyString(v["category"]),
			Brand:       anyString(v["brand"]),
			Currency:    anyString(v["currency"]),
			Images:      anyStringSlice(v["images"]),
			Tags:        anyStringSlice(v["tags"]),
		}
		if f, ok := anyFloat(v["price"]); ok {
			p.Price = f
		}
		if strings.TrimSpace(p.Title) == "" {
			return types.Product{}, fmt.Errorf("missing product.title")
		}
		return p, nil
	default:
		return types.Product{}, fmt.Errorf("missing product")
	}
}

func extractPublish(payload map[string]any) types.PublishRequest {
	if payload == nil {
		return types.PublishRequest{}
	}
	switch v := payload["publish"].(type) {
	case types.PublishRequest:
		return v
	case map[string]any:
		pub := types.PublishRequest{
			Platform:    anyString(v["platform"]),
			Channel:     anyString(v["channel"]),
			Operator:    anyString(v["operator"]),
			Content:     anyString(v["content"]),
			PublishMode: anyString(v["publish_mode"]),
			Page:        anyMap(v["page"]),
			Metadata:    anyMap(v["metadata"]),
		}
		return pub
	default:
		return types.PublishRequest{}
	}
}

func parseAlibabaResponse(resp AdapterHTTPResponse, action, requestID string) types.AdapterResponse {
	body := normalizeAlibabaResponse(resp.JSON)
	status := interpretAlibabaStatus(body, action)
	remoteID := firstNonEmpty(
		anyString(body["remote_id"]),
		anyString(body["item_id"]),
		anyString(body["id"]),
	)
	errCode := firstNonEmpty(anyString(body["error_code"]), anyString(body["sub_code"]), anyString(body["code"]))
	errMsg := firstNonEmpty(anyString(body["error_message"]), anyString(body["sub_msg"]), anyString(body["message"]))
	errClass := classifyAlibabaError(errCode, resp.StatusCode)
	if errCode != "" && errMsg == "" {
		errMsg = errCode
	}
	data := map[string]any{
		"http_status": resp.StatusCode,
		"response":    resp.JSON,
	}
	if remoteID != "" {
		data["remote_id"] = remoteID
	}
	if errCode != "" {
		data["error_code"] = errCode
	}
	if errClass != "" {
		data["error_class"] = errClass
	}
	if errMsg != "" {
		data["error_message"] = errMsg
	}
	return types.AdapterResponse{
		Platform:   "alibaba",
		Action:     action,
		Status:     status,
		Mode:       "live",
		RequestID:  requestID,
		Configured: true,
		Data:       data,
	}
}

func interpretAlibabaStatus(body map[string]any, action string) string {
	if success, ok := body["success"].(bool); ok {
		if success {
			return defaultStatusForAction(action)
		}
		return "failed"
	}
	if code := firstNonEmpty(anyString(body["error_code"]), anyString(body["sub_code"]), anyString(body["code"])); code != "" {
		return "failed"
	}
	return firstNonEmpty(anyString(body["status"]), anyString(body["result"]), defaultStatusForAction(action))
}

func classifyAlibabaError(code string, httpStatus int) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	switch {
	case httpStatus == 401 || strings.Contains(code, "AUTH") || strings.Contains(code, "SIGN"):
		return "auth"
	case httpStatus == 429 || strings.Contains(code, "LIMIT") || strings.Contains(code, "THROTTLE"):
		return "rate_limit"
	case httpStatus >= 500 || strings.Contains(code, "SYSTEM") || strings.Contains(code, "INTERNAL"):
		return "server"
	case code != "":
		return "business"
	default:
		return ""
	}
}

func normalizeAlibabaResponse(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	for _, key := range []string{"data", "result", "response"} {
		if nested, ok := m[key].(map[string]any); ok && len(nested) > 0 {
			return nested
		}
	}
	return m
}

func defaultStatusForAction(action string) string {
	switch action {
	case "publish":
		return "published"
	case "update":
		return "updated"
	case "on_shelf":
		return "on_shelf"
	case "off_shelf":
		return "off_shelf"
	default:
		return "success"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func anyString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func anyStringSlice(v any) []string {
	switch items := v.(type) {
	case []string:
		return items
	case []any:
		out := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func anyFloat(v any) (float64, bool) {
	f, ok := v.(float64)
	return f, ok
}

func anyMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
