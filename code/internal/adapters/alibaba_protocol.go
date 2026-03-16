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

type AlibabaLiveOptions struct {
	PublishPath    string
	UpdatePath     string
	OnShelfPath    string
	OffShelfPath   string
	SignMethod     string
	Version        string
	RequestFormat  string
	AuthPlacement  string
	IncludeAppKey  bool
	IncludeVersion bool
	MethodMap      map[string]string
}

func defaultAlibabaLiveOptions() AlibabaLiveOptions {
	return AlibabaLiveOptions{
		PublishPath:    "/publish",
		UpdatePath:     "/update",
		OnShelfPath:    "/on_shelf",
		OffShelfPath:   "/off_shelf",
		SignMethod:     config.GetEnv("JUYU_ALIBABA_SIGN_METHOD", "md5"),
		Version:        config.GetEnv("JUYU_ALIBABA_VERSION", "2.0"),
		RequestFormat:  config.GetEnv("JUYU_ALIBABA_REQUEST_FORMAT", "json"),
		AuthPlacement:  config.GetEnv("JUYU_ALIBABA_AUTH_PLACEMENT", "body"),
		IncludeAppKey:  !strings.EqualFold(config.GetEnv("JUYU_ALIBABA_INCLUDE_APP_KEY", "true"), "false"),
		IncludeVersion: !strings.EqualFold(config.GetEnv("JUYU_ALIBABA_INCLUDE_VERSION", "true"), "false"),
		MethodMap: map[string]string{
			"publish":   "alibaba.item.publish",
			"update":    "alibaba.item.update",
			"on_shelf":  "alibaba.item.on_shelf",
			"off_shelf": "alibaba.item.off_shelf",
		},
	}
}

func (a *AlibabaAdapter) liveOptions() AlibabaLiveOptions {
	return defaultAlibabaLiveOptions()
}

func (a *AlibabaAdapter) livePath(action string) string {
	opts := a.liveOptions()
	switch action {
	case "publish":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_PUBLISH_PATH", ""), opts.PublishPath)
	case "update":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_UPDATE_PATH", ""), opts.UpdatePath)
	case "on_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_ON_SHELF_PATH", ""), opts.OnShelfPath)
	case "off_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_OFF_SHELF_PATH", ""), opts.OffShelfPath)
	default:
		return "/" + action
	}
}

func (a *AlibabaAdapter) liveMethod(action string) string {
	switch action {
	case "publish":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_PUBLISH_METHOD", ""), a.liveOptions().MethodMap[action])
	case "update":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_UPDATE_METHOD", ""), a.liveOptions().MethodMap[action])
	case "on_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_ON_SHELF_METHOD", ""), a.liveOptions().MethodMap[action])
	case "off_shelf":
		return firstNonEmpty(config.GetEnv("JUYU_ALIBABA_OFF_SHELF_METHOD", ""), a.liveOptions().MethodMap[action])
	default:
		return a.liveOptions().MethodMap[action]
	}
}

func (a *AlibabaAdapter) signedPayload(action string, req types.AdapterRequest) (map[string]any, error) {
	creds, ok := a.credentialStore.Get(a.name)
	if !ok {
		return nil, fmt.Errorf("alibaba adapter credentials not found")
	}
	appKey := strings.TrimSpace(creds.Fields["app_key"])
	secret := strings.TrimSpace(creds.Fields["secret"])
	if appKey == "" || secret == "" {
		return nil, fmt.Errorf("alibaba adapter missing app_key or secret")
	}
	itemPayload, err := alibabaItemPayload(req)
	if err != nil {
		return nil, err
	}
	payloadJSON, err := json.Marshal(itemPayload)
	if err != nil {
		return nil, err
	}
	opts := a.liveOptions()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	params := map[string]string{
		"method":      a.liveMethod(action),
		"sign_method": opts.SignMethod,
		"timestamp":   timestamp,
		"payload":     string(payloadJSON),
	}
	if opts.IncludeAppKey {
		params["app_key"] = appKey
	}
	if opts.IncludeVersion {
		params["v"] = opts.Version
	}
	params["sign"] = alibabaMD5Sign(secret, params)
	return paramsWithItem(params, req, itemPayload), nil
}

func (a *AlibabaAdapter) liveInvoke(ctx context.Context, action string, req types.AdapterRequest) (types.AdapterResponse, error) {
	body, err := a.signedPayload(action, req)
	if err != nil {
		return types.AdapterResponse{}, err
	}
	parts := a.buildRequestParts(action, req, body)
	resp, err := a.httpClient.DoJSON(ctx, AdapterHTTPRequest{
		Method:      httpMethodForAction(action),
		URL:         parts.URL,
		Headers:     parts.Headers,
		Body:        parts.Body,
		ContentType: parts.ContentType,
		Timeout:     30 * time.Second,
	})
	if err != nil {
		return types.AdapterResponse{}, err
	}
	parsed := parseAlibabaResponse(resp, action, req.RequestID)
	parsed.Configured = a.isConfigured()
	parsed.Data["base_url"] = a.baseURL
	parsed.Data["external_ref"] = req.ExternalRef
	return parsed, nil
}

func alibabaMD5Sign(secret string, params map[string]string) string {
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
