package adapters

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryDoEventuallySucceeds(t *testing.T) {
	calls := 0
	out, err := retryDo(context.Background(), 3, time.Millisecond, func() (string, error) {
		calls++
		if calls < 3 {
			return "", errors.New("temporary")
		}
		return "ok", nil
	})
	if err != nil || out != "ok" || calls != 3 {
		t.Fatalf("unexpected retry result: out=%q err=%v calls=%d", out, err, calls)
	}
}
