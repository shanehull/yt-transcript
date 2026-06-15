package cache

import (
	"testing"
	"time"

	yt "github.com/shanehull/yt-transcript"
)

func TestGetSet(t *testing.T) {
	c := New(10 * time.Minute)
	segs := []yt.Segment{{Text: "hello", Start: 1.0, Duration: 2.0}}
	c.Set("key", segs)

	got, ok := c.Get("key")
	if !ok {
		t.Fatal("expected hit")
	}
	if len(got) != 1 || got[0].Text != "hello" {
		t.Fatalf("got %v", got)
	}
}

func TestExpiry(t *testing.T) {
	c := New(1 * time.Millisecond)
	c.Set("key", []yt.Segment{{Text: "x"}})
	time.Sleep(5 * time.Millisecond)
	if _, ok := c.Get("key"); ok {
		t.Fatal("expected miss after expiry")
	}
}

func TestMiss(t *testing.T) {
	c := New(10 * time.Minute)
	if _, ok := c.Get("nope"); ok {
		t.Fatal("expected miss")
	}
}
