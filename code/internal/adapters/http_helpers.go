package adapters

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func flattenFormValues(prefix string, v any, values url.Values) {
	switch x := v.(type) {
	case nil:
		return
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenFormValues(key, x[k], values)
		}
	case map[string]string:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			values.Set(key, x[k])
		}
	case []string:
		for i, item := range x {
			key := fmt.Sprintf("%s.%d", prefix, i)
			values.Set(strings.TrimPrefix(key, "."), item)
		}
	case []any:
		for i, item := range x {
			key := fmt.Sprintf("%s.%d", prefix, i)
			flattenFormValues(strings.TrimPrefix(key, "."), item, values)
		}
	default:
		values.Set(strings.TrimPrefix(prefix, "."), fmt.Sprint(v))
	}
}
