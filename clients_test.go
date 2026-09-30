package yt_transcript

import (
	"context"
	"testing"
)

func TestClientIdentityContext(t *testing.T) {
	ios := clientIdentity{name: "iOS", version: "20.11.6", deviceMake: "Apple", deviceModel: "iPhone10,4", osName: "iOS", osVersion: "16.7.7.20H330"}
	c := ios.context()
	if c["clientName"] != "iOS" || c["clientVersion"] != "20.11.6" {
		t.Fatalf("unexpected context: %v", c)
	}
	if _, ok := c["androidSdkVersion"]; ok {
		t.Error("androidSdkVersion should be omitted for a non-Android client")
	}

	android := clientIdentity{name: "ANDROID", version: "21.03.36", androidSDK: 36}
	if android.context()["androidSdkVersion"] != 36 {
		t.Error("androidSdkVersion should be set for Android")
	}
}

// TestLive_FetchEachClient is the rot tripwire: it exercises every configured
// identity, so a stale client shows up as a failure here rather than as a
// surprise in production.
func TestLive_FetchEachClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}
	c := NewClient()

	html, err := c.fetchWatchPage(context.Background(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("watch page: %v", err)
	}
	apiKey, err := extractAPIKey(html)
	if err != nil {
		t.Fatalf("api key: %v", err)
	}

	for _, id := range clientIdentities {
		segs, err := c.fetchWithClient(context.Background(), "dQw4w9WgXcQ", "en", apiKey, id)
		if err != nil {
			t.Errorf("%s: %v", id.name, err)
			continue
		}
		if len(segs) == 0 {
			t.Errorf("%s: no segments", id.name)
		}
	}
}
