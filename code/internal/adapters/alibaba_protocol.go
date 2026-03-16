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
	PublishPath  string
	UpdatePath   string
	OnShelfPath  string
	OffShelfPath string
	SignMethod   string
	Version      string
	MethodMap    map[string]string
}

func defaultAlibabaLiveOptions() AlibabaLiveOptions {
	return AlibabaLiveOptions{
		PublishPath:  "/publish",
		UpdatePath:   "/update",
		OnShelfPath:  "/on_shelf",
		OffShelfPath: "/off_shelf",
		SignMethod:   config.GetEnv("JUYU_ALIBABA_SIGN_METHOD", "md5"),
		Version:      config.GetEnv("JUYU_ALIBABA_VERSION", "2.0"),
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
		return opts.PublishPath
	case "update":
		return opts.UpdatePath
	case "on_shelf":
		return opts.OnShelfPath
	case "off_shelf":
		return opts.OffShelfPath
	default:
		return "/" + action
	}
}

func (a *AlibabaAdapter) liveMethod(action string) string {
	return a.liveOptions().MethodMap[action]
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
	payloadJSON, err := json.Marshal(requestPayload(req))
	if err != nil {
		return nil, err
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	params := map[string]string{
		"app_key":     appKey,
		"method":      a.liveMethod(action),
		"sign_method": a.liveOptions().SignMethod,
		"timestamp":   timestamp,
		"v":           a.liveOptions().Version,
		"payload":     string(payloadJSON),
	}
	params["sign"] = alibabaMD5Sign(secret, params)
	return map[string]any{
		"app_key":      params["app_key"],
		"method":       params["method"],
		"sign_method":  params["sign_method"],
		"timestamp":    params["timestamp"],
		"v":            params["v"],
		"sign":         params["sign"],
		"request_id":   req.RequestID,
		"external_ref": req.ExternalRef,
		"payload":      requestPayload(req),
	}, nil
}

func (a *AlibabaAdapter) liveInvoke(ctx context.Context, action string, req types.AdapterRequest) (types.AdapterResponse, error) {
	body, err := a.signedPayload(action, req)
	if err != nil {
		return types.AdapterResponse{}, err
	}
	return a.liveCall(ctx, action, req, a.livePath(action), body)
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
