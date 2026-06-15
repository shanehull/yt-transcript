// Package yt_transcript fetches YouTube video transcripts via the innertube API.
// Stdlib only, zero dependencies.
package yt_transcript

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"
)

const (
	watchURL  = "https://www.youtube.com/watch?v=%s"
	playerURL = "https://www.youtube.com/youtubei/v1/player?key=%s"
	apiKeyRe  = `"INNERTUBE_API_KEY":\s*"([a-zA-Z0-9_-]+)"`
	androidUA = "com.google.android.youtube/20.10.38 (Linux; U; Android 13) gzip"
)

// Segment represents a single transcript line with timing.
type Segment struct {
	Text     string  `json:"text"`
	Start    float64 `json:"start"`
	Duration float64 `json:"duration"`
}

// Client fetches YouTube transcripts.
type Client struct {
	http *http.Client
}

// NewClient returns a Client with a default 30s timeout and cookie jar.
func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second, Jar: jar},
	}
}

// WithHTTPClient sets a custom http.Client.
func (c *Client) WithHTTPClient(h *http.Client) *Client {
	c.http = h
	return c
}

// FetchTranscript retrieves the transcript for a YouTube video in the given language.
func (c *Client) FetchTranscript(ctx context.Context, videoID, lang string) ([]Segment, error) {
	html, err := c.fetchWatchPage(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("fetching watch page: %w", err)
	}

	apiKey, err := extractAPIKey(html)
	if err != nil {
		return nil, fmt.Errorf("YouTube did not return expected page content — the video may be unavailable or requests are being blocked: %w", err)
	}

	captions, err := c.fetchCaptions(ctx, videoID, apiKey)
	if err != nil {
		return nil, fmt.Errorf("fetching captions: %w", err)
	}

	transcript, err := findCaptionTrack(captions, lang)
	if err != nil {
		return nil, fmt.Errorf("finding transcript: %w", err)
	}

	return c.fetchAndParse(ctx, transcript.BaseURL)
}

func (c *Client) fetchWatchPage(ctx context.Context, videoID string) (string, error) {
	u := fmt.Sprintf(watchURL, videoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept-Language", "en-US")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("rate limited by YouTube (429)")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	html := string(body)

	if strings.Contains(html, `class="g-recaptcha"`) {
		return "", fmt.Errorf("bot detection triggered — YouTube is blocking requests from this IP")
	}

	return html, nil
}

var apiKeyRegex = regexp.MustCompile(apiKeyRe)

func extractAPIKey(html string) (string, error) {
	matches := apiKeyRegex.FindStringSubmatch(html)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse YouTube page")
	}
	return matches[1], nil
}

type captionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Kind         string `json:"kind"`
	Name         struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"name"`
}

func (c *Client) fetchCaptions(ctx context.Context, videoID, apiKey string) ([]captionTrack, error) {
	body, err := json.Marshal(map[string]any{
		"context": map[string]any{
			"client": map[string]string{
				"clientName":    "ANDROID",
				"clientVersion": "20.10.38",
			},
		},
		"videoId": videoID,
	})
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(playerURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", androidUA)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited by YouTube (429)")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Captions struct {
			PlayerCaptionsTracklistRenderer struct {
				CaptionTracks []captionTrack `json:"captionTracks"`
			} `json:"playerCaptionsTracklistRenderer"`
		} `json:"captions"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("parsing player response: %w", err)
	}

	tracks := data.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no caption tracks available")
	}
	return tracks, nil
}

func findCaptionTrack(tracks []captionTrack, lang string) (captionTrack, error) {
	for _, t := range tracks {
		if t.LanguageCode == lang {
			t.BaseURL = strings.Replace(t.BaseURL, "&fmt=srv3", "", 1)
			return t, nil
		}
	}
	return captionTrack{}, fmt.Errorf("no transcript found for language %q", lang)
}

type transcriptXML struct {
	XMLName xml.Name         `xml:"transcript"`
	Texts   []transcriptText `xml:"text"`
}

type transcriptText struct {
	Start float64 `xml:"start,attr"`
	Dur   float64 `xml:"dur,attr"`
	Value string  `xml:",chardata"`
}

func (c *Client) fetchAndParse(ctx context.Context, baseURL string) ([]Segment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited by YouTube (429)")
	}

	var x transcriptXML
	if err := xml.NewDecoder(resp.Body).Decode(&x); err != nil {
		return nil, fmt.Errorf("parsing transcript XML: %w", err)
	}

	segments := make([]Segment, len(x.Texts))
	for i, t := range x.Texts {
		segments[i] = Segment{
			Text:     unescapeHTML(t.Value),
			Start:    t.Start,
			Duration: t.Dur,
		}
	}
	return segments, nil
}

var htmlReplacer = strings.NewReplacer(
	"&amp;", "&",
	"&lt;", "<",
	"&gt;", ">",
	"&quot;", `"`,
	"&#39;", "'",
	"&apos;", "'",
)

func unescapeHTML(s string) string {
	return htmlReplacer.Replace(s)
}
