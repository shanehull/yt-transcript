// Package handlers provides HTTP handlers for the yt-transcript server.
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	yt "github.com/shanehull/yt-transcript"
	"github.com/shanehull/yt-transcript/internal/cache"
	"github.com/shanehull/yt-transcript/internal/middleware"
)

var faviconSVG = []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#14b8a6"/><text x="16" y="22" font-family="system-ui,sans-serif" font-size="16" font-weight="700" fill="#fff" text-anchor="middle">YT</text></svg>`)

var faviconDataURI = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(faviconSVG)

var indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>yt-transcript — fetch YouTube transcripts, no API key required</title>
<meta name="description" content="Fetch YouTube video transcripts as JSON, plain text, or SRT. No API key or signup. Ideal for piping into LLMs, indexing video content, or generating subtitles.">
<meta property="og:title" content="yt-transcript — fetch YouTube transcripts">
<meta property="og:description" content="Fetch YouTube transcripts as JSON, text, or SRT. No API key. Great for LLMs and automation.">
<meta property="og:type" content="website">
<meta property="og:url" content="{{.BaseURL}}">
<link rel="icon" href="` + faviconDataURI + `">
<style>
  *,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
  :root{color-scheme:dark}
  body{font-family:system-ui,-apple-system,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;background:#09090b;color:#a1a1aa;line-height:1.6;-webkit-font-smoothing:antialiased}
  .page{max-width:860px;margin:0 auto;padding:4rem 1.5rem 3rem}
  .hero{margin-bottom:3rem}
  .hero h1{font-size:2rem;font-weight:700;color:#fafafa;letter-spacing:-.02em}
  .hero p{color:#52525b;margin-top:.5rem;font-size:1.05rem;line-height:1.7}
  .section{margin-bottom:2.5rem}
  .section-title{font-size:.7rem;font-weight:600;color:#3f3f46;text-transform:uppercase;letter-spacing:.08em;margin-bottom:.75rem}
  .card{background:#111113;border:1px solid #1f1f23;border-radius:8px;overflow:hidden}
  .card-body{padding:1rem}
  .card-body pre{margin:0;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,monospace;font-size:.82rem;line-height:1.65;color:#a1a1aa;white-space:pre-wrap;word-break:break-word}
  .card-body .hl{color:#2dd4bf}
  .card-body .url{color:#60a5fa}
  .card-body .dim{color:#3f3f46}
  .badge{display:inline-block;font-size:.65rem;font-weight:600;color:#2dd4bf;background:#2dd4bf12;padding:.15em .5em;border-radius:4px;vertical-align:middle;margin-left:.4rem}
  .footer{margin-top:3rem;padding-top:1.5rem;border-top:1px solid #1f1f23;color:#3f3f46;font-size:.8rem}
  .footer a{color:#52525b;text-decoration:none}
  .footer a:hover{color:#a1a1aa}
  .footer .version{float:right;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,monospace}
  .endpoint{font-size:.82rem;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,monospace}
  .endpoint .method{color:#2dd4bf;font-weight:600}
  .endpoint .path{color:#fafafa}
</style>
</head>
<body>
<div class="page">
  <div class="hero">
    <h1>yt-transcript</h1>
    <p>Fetch the transcript of any YouTube video as JSON, plain text, or SRT. If you're piping transcripts into an LLM, indexing video content, or generating subtitles, this is the fastest way to get captions out of YouTube.</p>
  </div>

  <div class="section">
    <div class="section-title">Endpoint</div>
    <div class="card">
      <div class="card-body">
        <div class="endpoint"><span class="method">GET</span> <span class="path">/{video_id}</span> <span class="dim">— transcript</span></div>
      </div>
    </div>
  </div>

  <div class="section">
    <div class="section-title">Quick start</div>
    <div class="card">
      <div class="card-body">
        <pre><span class="hl">curl</span> <span class="url">{{.BaseURL}}</span>/dQw4w9WgXcQ</pre>
        <pre><span class="hl">curl</span> <span class="url">{{.BaseURL}}</span>/dQw4w9WgXcQ?<span class="hl">fmt=</span>text</pre>
        <pre><span class="hl">curl</span> <span class="url">{{.BaseURL}}</span>/dQw4w9WgXcQ?<span class="hl">fmt=</span>srt</pre>
        <pre><span class="hl">curl</span> <span class="url">{{.BaseURL}}</span>/dQw4w9WgXcQ?<span class="hl">lang=</span>fr</pre>
      </div>
    </div>
  </div>

  <div class="section">
    <div class="section-title">Response formats <span class="badge">?fmt=</span></div>
    <div class="card">
      <div class="card-body">
        <div class="section-title" style="margin-bottom:.5rem">json <span class="badge">default</span></div>
        <pre><span class="dim">[</span>
  <span class="dim">{</span> <span class="hl">&quot;text&quot;</span>: <span class="url">&quot;We&apos;re no strangers to love&quot;</span><span class="dim">,</span> <span class="hl">&quot;start&quot;</span>: 18.53<span class="dim">,</span> <span class="hl">&quot;duration&quot;</span>: 3.2 <span class="dim">}</span><span class="dim">,</span>
  <span class="dim">{</span> <span class="hl">&quot;text&quot;</span>: <span class="url">&quot;You know the rules and so do I&quot;</span><span class="dim">,</span> <span class="hl">&quot;start&quot;</span>: 21.73<span class="dim">,</span> <span class="hl">&quot;duration&quot;</span>: 2.82 <span class="dim">}</span>
<span class="dim">]</span></pre>
      </div>
    </div>

    <div class="card" style="margin-top:.75rem">
      <div class="card-body">
        <div class="section-title" style="margin-bottom:.5rem">text</div>
        <pre>We&apos;re no strangers to love
You know the rules and so do I</pre>
      </div>
    </div>

    <div class="card" style="margin-top:.75rem">
      <div class="card-body">
        <div class="section-title" style="margin-bottom:.5rem">srt</div>
        <pre><span class="hl">1</span>
<span class="url">00:00:18,530</span> <span class="dim">--&gt;</span> <span class="url">00:00:21,730</span>
We&apos;re no strangers to love

<span class="hl">2</span>
<span class="url">00:00:21,730</span> <span class="dim">--&gt;</span> <span class="url">00:00:24,550</span>
You know the rules and so do I</pre>
      </div>
    </div>
  </div>

  <div class="footer">
    <a href="https://github.com/shanehull/yt-transcript">github.com/shanehull/yt-transcript</a>
    <span class="version">{{.Version}}</span>
  </div>
</div>
</body>
</html>`

type pageData struct {
	BaseURL string
	Version string
}

var videoIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

var indexPage = template.Must(template.New("index").Parse(indexHTML))

// Healthz returns a 200 OK status for health checks.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Favicon serves the SVG favicon.
func Favicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(faviconSVG)
}

// Index serves the landing page with usage examples.
func Index(baseURL, version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = indexPage.Execute(w, pageData{BaseURL: baseURL, Version: version})
	})
}

// Transcript fetches and serves a YouTube video transcript.
func Transcript(client *yt.Client, transcriptCache *cache.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videoID := r.PathValue("video_id")
		if videoID == "" || !videoIDRe.MatchString(videoID) {
			writeError(w, http.StatusBadRequest, "invalid video_id")
			return
		}

		lang := r.URL.Query().Get("lang")
		if lang == "" {
			lang = "en"
		}

		cacheKey := videoID + ":" + lang

		var segments []yt.Segment
		var cacheHit bool
		if segs, ok := transcriptCache.Get(cacheKey); ok {
			segments = segs
			cacheHit = true
		} else {
			var err error
			segments, err = client.FetchTranscript(r.Context(), videoID, lang)
			if err != nil {
				slog.Error("fetching transcript", "video_id", videoID, "lang", lang, "error", err)
				writeError(w, errorStatus(err), err.Error())
				return
			}
			transcriptCache.Set(cacheKey, segments)
		}

		if cs, ok := r.Context().Value(middleware.CtxKey{}).(*middleware.CacheStatus); ok {
			cs.Hit = cacheHit
			cs.Set = true
		}

		writeSegments(w, r.URL.Query().Get("fmt"), segments)
	})
}

func writeSegments(w http.ResponseWriter, format string, segments []yt.Segment) {
	switch format {
	case "text":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, s := range segments {
			_, _ = fmt.Fprintln(w, s.Text)
		}
	case "srt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for i, s := range segments {
			_, _ = fmt.Fprintf(w, "%d\n%s --> %s\n%s\n\n",
				i+1, srtTime(s.Start), srtTime(s.Start+s.Duration),
				strings.TrimSpace(s.Text))
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(segments)
	}
}

func srtTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "rate limited"):
		return http.StatusServiceUnavailable
	case strings.Contains(msg, "bot detection"):
		return http.StatusServiceUnavailable
	case strings.Contains(msg, "could not parse YouTube page"):
		return http.StatusBadGateway
	case strings.Contains(msg, "no caption tracks"),
		strings.Contains(msg, "no transcript found"):
		return http.StatusNotFound
	default:
		return http.StatusBadGateway
	}
}
