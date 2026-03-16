package adapters

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"juyu-ai-platform/internal/config"
	"juyu-ai-platform/internal/types"
)

type TaobaoLiveOptions struct {
	PublishPath   string
	UpdatePath    string
	OnShelfPath   string
	OffShelfPath  string
	SignMethod    string
	Version       string
	RequestFormat string
	AuthPlacement string
	MethodMap     map[string]string
}

func defaultTaobaoLiveOptions() TaobaoLiveOptions {
	return TaobaoLiveOptions{
		PublishPath:   "/publish",
		UpdatePath:    "/update",
		OnShelfPath:   "/on_shelf",
		OffShelfPath:  "/off_shelf",
		SignMethod:    config.GetEnv("JUYU_TAOBAO_SIGN_METHOD", "md5"),
		Version:       config.GetEnv("JUYU_TAOBAO_VERSION", "2.0"),
		RequestFormat: config.GetEnv("JUYU_TAOBAO_REQUEST_FORMAT", "json"),
		AuthPlacement: config.GetEnv("JUYU_TAOBAO_AUTH_PLACEMENT", "body"),
		MethodMap: map[string]string{
			"publish":   "taobao.item.publish",
			"update":    "taobao.item.update",
			"on_shelf":  "taobao.item.on_shelf",
			"off_shelf": "taobao.item.off_shelf",
		},
	}
}

func (a *TaobaoAdapter) liveOptions() TaobaoLiveOptions { return defaultTaobaoLiveOptions() }

func (a *TaobaoAdapter) livePath(action string) string {
	switch action {
	case "publish":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_PUBLISH_PATH", ""), a.liveOptions().PublishPath)
	case "update":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_UPDATE_PATH", ""), a.liveOptions().UpdatePath)
	case "on_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_ON_SHELF_PATH", ""), a.liveOptions().OnShelfPath)
	case "off_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_OFF_SHELF_PATH", ""), a.liveOptions().OffShelfPath)
	default:
		return "/" + action
	}
}

func (a *TaobaoAdapter) liveMethod(action string) string {
	switch action {
	case "publish":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_PUBLISH_METHOD", ""), a.liveOptions().MethodMap[action])
	case "update":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_UPDATE_METHOD", ""), a.liveOptions().MethodMap[action])
	case "on_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_ON_SHELF_METHOD", ""), a.liveOptions().MethodMap[action])
	case "off_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_TAOBAO_OFF_SHELF_METHOD", ""), a.liveOptions().MethodMap[action])
	default:
		return a.liveOptions().MethodMap[action]
	}
}

func (a *TaobaoAdapter) signedPayload(action string, req types.AdapterRequest) (map[string]any, error) {
	creds, ok := a.credentialStore.Get(a.name)
	if !ok {
		return nil, fmt.Errorf("taobao adapter credentials not found")
	}
	appKey := strings.TrimSpace(creds.Fields["app_key"])
	secret := strings.TrimSpace(creds.Fields["secret"])
	if appKey == "" || secret == "" {
		return nil, fmt.Errorf("taobao adapter missing app_key or secret")
	}
	itemPayload, err := alibabaItemPayload(req)
	if err != nil {
		return nil, err
	}
	payloadJSON, err := json.Marshal(itemPayload)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"app_key":     appKey,
		"method":      a.liveMethod(action),
		"sign_method": a.liveOptions().SignMethod,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"v":           a.liveOptions().Version,
		"payload":     string(payloadJSON),
	}
	params["sign"] = taobaoMD5Sign(secret, params)
	return paramsWithItem(params, req, itemPayload), nil
}

func (a *TaobaoAdapter) buildRequestParts(action string, req types.AdapterRequest, body map[string]any) AdapterRequestParts {
	parts := AdapterRequestParts{URL: a.endpoint(a.livePath(action)), Headers: map[string]string{}, Body: body, ContentType: "application/json"}
	if strings.EqualFold(a.liveOptions().RequestFormat, "form") || strings.EqualFold(a.liveOptions().RequestFormat, "x-www-form-urlencoded") {
		parts.ContentType = "application/x-www-form-urlencoded"
	}
	parts.Headers["X-Request-ID"] = req.RequestID
	parts.Headers["X-External-Ref"] = req.ExternalRef
	if strings.EqualFold(a.liveOptions().AuthPlacement, "query") {
		query := map[string]string{}
		for _, key := range []string{"app_key", "method", "sign_method", "timestamp", "v", "sign"} {
			if s, ok := body[key].(string); ok && s != "" {
				query[key] = s
				delete(body, key)
			}
		}
		parts.URL = buildRequestURL(parts.URL, query)
	}
	return parts
}

func (a *TaobaoAdapter) liveInvoke(ctx context.Context, action string, req types.AdapterRequest) (types.AdapterResponse, error) {
	body, err := a.signedPayload(action, req)
	if err != nil {
		return types.AdapterResponse{}, err
	}
	parts := a.buildRequestParts(action, req, body)
	resp, err := a.httpClient.DoJSON(ctx, AdapterHTTPRequest{Method: httpMethodForAction(action), URL: parts.URL, Headers: parts.Headers, Body: parts.Body, ContentType: parts.ContentType, Timeout: 30 * time.Second})
	if err != nil {
		return types.AdapterResponse{}, err
	}
	parsed := parseAlibabaResponse(resp, action, req.RequestID)
	parsed.Platform = "taobao"
	parsed.Configured = a.isConfigured()
	parsed.Data["base_url"] = a.baseURL
	return parsed, nil
}

func taobaoMD5Sign(secret string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if strings.TrimSpace(k) == "" || k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(secret)
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(params[k])
	}
	b.WriteString(secret)
	sum := md5.Sum([]byte(b.String()))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}
