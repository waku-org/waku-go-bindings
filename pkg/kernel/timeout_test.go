package kernel

import (
	"context"
	"testing"
	"time"
)

// Zero milliseconds is not "no timeout": chronos expires on it immediately, so
// a context without a deadline has to fall back to the request timeout.
func TestContextTimeoutFallsBackToRequestTimeout(t *testing.T) {
	if got, want := getContextTimeoutMilliseconds(context.Background()),
		int(requestTimeout.Milliseconds()); got != want {
		t.Errorf("getContextTimeoutMilliseconds(Background) = %d, want %d", got, want)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if got := getContextTimeoutMilliseconds(ctx); got <= 0 || got > 60_000 {
		t.Errorf("getContextTimeoutMilliseconds(1m) = %d, want (0, 60000]", got)
	}

	expired, cancelExpired := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelExpired()
	if got := getContextTimeoutMilliseconds(expired); got != 0 {
		t.Errorf("getContextTimeoutMilliseconds(expired) = %d, want 0", got)
	}
}
