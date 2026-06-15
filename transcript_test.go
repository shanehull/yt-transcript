package yt_transcript

import (
	"context"
	"testing"
)

func TestLive_FetchTranscript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}
	c := NewClient()

	segs, err := c.FetchTranscript(context.Background(), "dQw4w9WgXcQ", "en")
	if err != nil {
		t.Fatalf("FetchTranscript: %v", err)
	}
	if len(segs) == 0 {
		t.Fatal("expected >0 segments")
	}
	for _, s := range segs {
		if s.Text == "" {
			t.Error("empty segment text")
		}
		if s.Start < 0 {
			t.Errorf("negative start time: %f", s.Start)
		}
	}
}

func TestLive_FetchTranscriptMissingLang(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}
	c := NewClient()
	_, err := c.FetchTranscript(context.Background(), "dQw4w9WgXcQ", "xx")
	if err == nil {
		t.Fatal("expected error for missing language")
	}
}

func TestLive_FetchTranscriptMissingVideo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}
	c := NewClient()
	_, err := c.FetchTranscript(context.Background(), "nonexistent_video_id_12345", "en")
	if err == nil {
		t.Fatal("expected error for missing video")
	}
}

func TestExtractAPIKey(t *testing.T) {
	html := `<html><script>"INNERTUBE_API_KEY": "test_api_key_12345"</script></html>`
	key, err := extractAPIKey(html)
	if err != nil {
		t.Fatalf("extractAPIKey: %v", err)
	}
	if key != "test_api_key_12345" {
		t.Fatalf("got %q, want %q", key, "test_api_key_12345")
	}
}

func TestExtractAPIKeyMissing(t *testing.T) {
	_, err := extractAPIKey("<html></html>")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindCaptionTrack(t *testing.T) {
	tracks := []captionTrack{
		{LanguageCode: "en", BaseURL: "https://example.com/en?key=v&fmt=srv3&other=x"},
		{LanguageCode: "fr", BaseURL: "https://example.com/fr"},
	}
	track, err := findCaptionTrack(tracks, "en")
	if err != nil {
		t.Fatalf("findCaptionTrack: %v", err)
	}
	if track.LanguageCode != "en" {
		t.Fatalf("got %q, want en", track.LanguageCode)
	}
	if track.BaseURL != "https://example.com/en?key=v&other=x" {
		t.Fatalf("fmt=srv3 not stripped: %q", track.BaseURL)
	}
}

func TestFindCaptionTrackNoFmt(t *testing.T) {
	tracks := []captionTrack{
		{LanguageCode: "fr", BaseURL: "https://example.com/fr"},
	}
	track, err := findCaptionTrack(tracks, "fr")
	if err != nil {
		t.Fatalf("findCaptionTrack: %v", err)
	}
	if track.BaseURL != "https://example.com/fr" {
		t.Fatalf("unexpected URL modification: %q", track.BaseURL)
	}
}

func TestFindCaptionTrackMissing(t *testing.T) {
	_, err := findCaptionTrack(nil, "zz")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnescapeHTML(t *testing.T) {
	cases := map[string]string{
		"hello &amp; world":      "hello & world",
		"x &lt; y":               "x < y",
		"he said &quot;hi&quot;": `he said "hi"`,
		"that&#39;s cool":        "that's cool",
	}
	for in, want := range cases {
		if got := unescapeHTML(in); got != want {
			t.Errorf("unescapeHTML(%q) = %q, want %q", in, got, want)
		}
	}
}
