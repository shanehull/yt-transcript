# yt-transcript

[![Go Reference](https://pkg.go.dev/badge/github.com/shanehull/yt-transcript.svg)](https://pkg.go.dev/github.com/shanehull/yt-transcript)
[![Go Report Card](https://goreportcard.com/badge/github.com/shanehull/yt-transcript)](https://goreportcard.com/report/github.com/shanehull/yt-transcript)
[![CI](https://github.com/shanehull/yt-transcript/actions/workflows/test.yaml/badge.svg)](https://github.com/shanehull/yt-transcript/actions/workflows/test.yaml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Fetch YouTube transcripts. No API key needed.

## Install

```bash
go install github.com/shanehull/yt-transcript/cmd/yt-transcript@latest
```

## CLI

```bash
yt-transcript dQw4w9WgXcQ                  # plain text
yt-transcript -fmt json dQw4w9WgXcQ        # JSON with timestamps
yt-transcript -fmt srt dQw4w9WgXcQ         # SRT subtitles
yt-transcript -lang fr dQw4w9WgXcQ         # French transcript
```

## Library

```go
import (
    "context"
    yt "github.com/shanehull/yt-transcript"
)

client := yt.NewClient()
segments, _ := client.FetchTranscript(context.Background(), "dQw4w9WgXcQ", "en")
for _, s := range segments {
    fmt.Println(s.Text)
}
```

## Server

```bash
go run github.com/shanehull/yt-transcript/cmd/server@latest
docker run -p 8080:8080 ghcr.io/shanehull/yt-transcript
```

```
GET /{video_id}[?lang=en][&fmt=text|srt]  → transcript
```

```bash
curl https://yt-transcript.net/dQw4w9WgXcQ
curl https://yt-transcript.net/dQw4w9WgXcQ?fmt=text
curl https://yt-transcript.net/dQw4w9WgXcQ?fmt=srt
curl https://yt-transcript.net/dQw4w9WgXcQ?lang=fr
```
