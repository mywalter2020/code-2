package adapters

import (
	"fmt"
	"strconv"
	"strings"
)

func parsePositiveInt(s string) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid positive int: %q", s)
	}
	return v, nil
}

func mapStringAny(m map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	return out
}
