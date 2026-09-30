package yt_transcript

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBreakerTripsAfterThreshold(t *testing.T) {
	b := newBreaker(3, time.Minute, time.Hour)
	now := time.Unix(0, 0)

	for i := 0; i < 2; i++ {
		if d := b.fail(0, now); d != 0 {
			t.Fatalf("fail %d: cooldown = %s, want 0", i, d)
		}
		if _, ok := b.allow(now); !ok {
			t.Fatalf("fail %d: breaker open below threshold", i)
		}
	}

	if d := b.fail(0, now); d != time.Minute {
		t.Fatalf("third fail: cooldown = %s, want 1m", d)
	}
	remaining, ok := b.allow(now)
	if ok || remaining != time.Minute {
		t.Fatalf("allow after trip = (%s, %v), want (1m, false)", remaining, ok)
	}
}

func TestBreakerCooldownDoubles(t *testing.T) {
	b := newBreaker(1, time.Minute, time.Hour)
	now := time.Unix(0, 0)

	if d := b.fail(0, now); d != time.Minute {
		t.Fatalf("first open = %s, want 1m", d)
	}
	now = now.Add(time.Minute)
	if _, ok := b.allow(now); !ok {
		t.Fatal("expected probe once cooldown elapsed")
	}
	if d := b.fail(0, now); d != 2*time.Minute {
		t.Fatalf("second open = %s, want 2m", d)
	}
}

func TestBreakerCapsCooldown(t *testing.T) {
	b := newBreaker(1, time.Minute, 5*time.Minute)
	now := time.Unix(0, 0)
	for i := 0; i < 10; i++ {
		b.fail(0, now)
	}
	if d := b.fail(0, now); d != 5*time.Minute {
		t.Fatalf("cooldown = %s, want capped 5m", d)
	}
}

func TestBreakerRespectsRetryAfter(t *testing.T) {
	b := newBreaker(1, time.Minute, time.Hour)
	if d := b.fail(30*time.Second, time.Unix(0, 0)); d != 30*time.Second {
		t.Fatalf("open = %s, want 30s", d)
	}
}

func TestBreakerSuccessResets(t *testing.T) {
	b := newBreaker(2, time.Minute, time.Hour)
	now := time.Unix(0, 0)
	b.fail(0, now)
	b.succeed()
	if d := b.fail(0, now); d != 0 {
		t.Fatalf("after success: cooldown = %s, want 0", d)
	}
	if d := b.fail(0, now); d != time.Minute {
		t.Fatalf("second consecutive fail = %s, want 1m", d)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func stubClient(fn roundTripFunc) *Client {
	return NewClient().WithHTTPClient(&http.Client{Transport: fn})
}

func TestClientTripsBreakerOn429(t *testing.T) {
	var calls int
	client := stubClient(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})

	for i := 0; i < 3; i++ {
		if _, err := client.fetchWatchPage(context.Background(), "dQw4w9WgXcQ"); err == nil {
			t.Fatalf("call %d: expected rate limit error", i)
		}
	}
	if calls != 3 {
		t.Fatalf("transport calls = %d, want 3", calls)
	}

	_, err := client.fetchWatchPage(context.Background(), "dQw4w9WgXcQ")
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("error = %v, want *RateLimitError", err)
	}
	if rle.RetryAfter <= 0 {
		t.Fatalf("RetryAfter = %s, want > 0", rle.RetryAfter)
	}
	if calls != 3 {
		t.Fatalf("breaker did not short-circuit: transport calls = %d, want 3", calls)
	}
}

func TestClientHonoursRetryAfterHeader(t *testing.T) {
	client := stubClient(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{"42"}},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})

	_, err := client.fetchWatchPage(context.Background(), "dQw4w9WgXcQ")
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("error = %v, want *RateLimitError", err)
	}
	if rle.RetryAfter != 42*time.Second {
		t.Fatalf("RetryAfter = %s, want 42s", rle.RetryAfter)
	}
}

func TestClientResetsBreakerOnSuccess(t *testing.T) {
	var calls int
	client := stubClient(func(*http.Request) (*http.Response, error) {
		calls++
		if calls <= 3 {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("<html>x</html>")),
		}, nil
	})

	for i := 0; i < 3; i++ {
		_, _ = client.fetchWatchPage(context.Background(), "dQw4w9WgXcQ")
	}
	if _, ok := client.breaker.allow(time.Now()); ok {
		t.Fatal("breaker should be open after three 429s")
	}

	// Elapse the cooldown so the probe is admitted, then succeed.
	client.breaker.mu.Lock()
	client.breaker.openUntil = time.Now().Add(-time.Second)
	client.breaker.mu.Unlock()

	if _, err := client.fetchWatchPage(context.Background(), "dQw4w9WgXcQ"); err != nil {
		t.Fatalf("probe: %v", err)
	}
	if _, ok := client.breaker.allow(time.Now()); !ok {
		t.Fatal("breaker should close after success")
	}
}
