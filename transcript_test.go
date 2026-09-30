package yt_transcript

import (
	"context"
	"reflect"
	"testing"
)

func langTrack(code, kind, name string) captionTrack {
	t := captionTrack{LanguageCode: code, Kind: kind}
	if name != "" {
		t.Name.Runs = append(t.Name.Runs, struct {
			Text string `json:"text"`
		}{Text: name})
	}
	return t
}

func TestToLanguages(t *testing.T) {
	got := toLanguages([]captionTrack{
		langTrack("en", "asr", "English (auto-generated)"),
		langTrack("en", "", "English"),
		langTrack("de", "", "German"),
		langTrack("fr", "", ""),
	})
	want := []TranscriptLanguage{
		{Code: "en", Name: "English"},
		{Code: "de", Name: "German"},
		{Code: "fr", Name: "fr"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("toLanguages = %+v, want %+v", got, want)
	}
}

func TestLive_ListLanguages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}
	langs, err := NewClient().ListLanguages(context.Background(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("ListLanguages: %v", err)
	}
	if len(langs) == 0 {
		t.Fatal("expected at least one language")
	}
	var found bool
	for _, l := range langs {
		if l.Code == "" {
			t.Error("empty language code")
		}
		if l.Name == "" {
			t.Errorf("%s: empty language name", l.Code)
		}
		if l.Code == "en" {
			found = true
		}
	}
	if !found {
		t.Error("expected English in the language list")
	}
}

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
