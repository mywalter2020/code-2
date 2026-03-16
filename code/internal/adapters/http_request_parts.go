package adapters

import "net/url"

type AdapterRequestParts struct {
	URL         string
	Headers     map[string]string
	Body        any
	ContentType string
}

func buildRequestURL(rawURL string, query map[string]string) string {
	if len(query) == 0 {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for k, v := range query {
		if k == "" || v == "" {
			continue
		}
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
