package types

import "strings"

// Product 表示商品基础业务对象。
type Product struct {
	ID          string   `json:"id,omitempty"`
	SKU         string   `json:"sku,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Brand       string   `json:"brand,omitempty"`
	Price       float64  `json:"price,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	Images      []string `json:"images,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// PublishRequest 表示一次发布动作的业务载荷。
type PublishRequest struct {
	Platform    string         `json:"platform"`
	Channel     string         `json:"channel,omitempty"`
	Operator    string         `json:"operator,omitempty"`
	Product     Product        `json:"product"`
	Content     string         `json:"content,omitempty"`
	Page        map[string]any `json:"page,omitempty"`
	PublishMode string         `json:"publish_mode,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ConfirmationPayload 表示人工确认动作的业务对象。
type ConfirmationPayload struct {
	TaskID         string         `json:"task_id,omitempty"`
	Required       bool           `json:"required"`
	Status         string         `json:"status,omitempty"`
	Reviewer       string         `json:"reviewer,omitempty"`
	Comment        string         `json:"comment,omitempty"`
	PreviewSummary string         `json:"preview_summary,omitempty"`
	Checklist      []string       `json:"checklist,omitempty"`
	Snapshot       map[string]any `json:"snapshot,omitempty"`
}

// BusinessContext 将通用 Request 归一化为业务对象模型，便于前后端对齐。
type BusinessContext struct {
	Scene        string              `json:"scene"`
	Input        string              `json:"input"`
	Operator     string              `json:"operator,omitempty"`
	Product      Product             `json:"product"`
	Publish      PublishRequest      `json:"publish"`
	Confirmation ConfirmationPayload `json:"confirmation"`
}

func BuildBusinessContext(req Request, results []Response) BusinessContext {
	ctx := BusinessContext{
		Scene:    req.Scene,
		Input:    req.Input,
		Operator: req.Operator,
		Product: Product{
			Title:       req.Input,
			Description: req.Input,
			Currency:    stringFromMap(req.Payload, "currency", "CNY"),
			Category:    stringFromMap(req.Payload, "category", "default"),
			Brand:       stringFromMap(req.Payload, "brand", "JuYu"),
			SKU:         stringFromMap(req.Payload, "sku", "DEMO-SKU-001"),
			Tags:        stringSliceFromAny(req.Payload["tags"]),
			Images:      stringSliceFromAny(req.Payload["images"]),
		},
		Publish: PublishRequest{
			Platform:    stringFromMap(req.Payload, "platform", "alibaba"),
			Channel:     stringFromMap(req.Payload, "channel", "default"),
			Operator:    req.Operator,
			PublishMode: stringFromMap(req.Payload, "publish_mode", "manual_confirm"),
			Metadata:    map[string]any{},
		},
		Confirmation: ConfirmationPayload{
			Required: true,
			Status:   "pending",
			Checklist: []string{
				"商品标题与卖点已确认",
				"价格与平台已确认",
				"发布前预览已确认",
			},
			Snapshot: map[string]any{},
		},
	}

	if req.Payload != nil {
		if title := stringFromMap(req.Payload, "title", ""); title != "" {
			ctx.Product.Title = title
		}
		if subtitle := stringFromMap(req.Payload, "subtitle", ""); subtitle != "" {
			ctx.Product.Subtitle = subtitle
		}
		if desc := stringFromMap(req.Payload, "description", ""); desc != "" {
			ctx.Product.Description = desc
		}
		if price, ok := req.Payload["price"].(float64); ok {
			ctx.Product.Price = price
		}
	}

	ctx.Publish.Product = ctx.Product

	for _, result := range results {
		switch result.Agent {
		case "content_gen":
			ctx.Publish.Content = stringFromData(result.Data, "content")
			ctx.Publish.Metadata["content_summary"] = stringFromData(result.Data, "summary")
			ctx.Confirmation.Snapshot["content"] = result.Data
		case "page_gen":
			if page, ok := result.Data["page"].(map[string]any); ok {
				ctx.Publish.Page = page
			}
			ctx.Confirmation.Snapshot["page"] = result.Data
		case "review_check":
			ctx.Confirmation.PreviewSummary = stringFromData(result.Data, "summary")
			ctx.Confirmation.Status = "awaiting_review"
			ctx.Confirmation.Snapshot["review"] = result.Data
		case "publish_exec":
			ctx.Publish.Metadata["publish_status"] = result.Data["publish_status"]
			ctx.Confirmation.Snapshot["publish"] = result.Data
		case "onshelf_exec":
			ctx.Publish.Metadata["shelf_status"] = result.Data["shelf_status"]
			ctx.Confirmation.Snapshot["shelf"] = result.Data
		}
	}

	return ctx
}

func stringFromMap(m map[string]any, key, fallback string) string {
	if m == nil {
		return fallback
	}
	if v, ok := m[key].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func stringFromData(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func stringSliceFromAny(v any) []string {
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
