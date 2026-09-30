---
name: yt-transcript
description: Fetch the transcript of a YouTube video, or list the caption languages it offers. Use this skill when the user wants captions, subtitles, the full text of what was said in a YouTube video, or the languages a video is available in.
compatibility: Requires network access.
allowed-tools: bash
---

# yt-transcript

```
GET https://yt-transcript.net/{video_id}[?lang=en][&fmt=text|json|srt]
GET https://yt-transcript.net/{video_id}/languages
```

Defaults to English, JSON output.

## Gotchas

- **Video ID only** — pass `dQw4w9WgXcQ`, not a URL. Strip `https://www.youtube.com/watch?v=` and any `&t=` or `&list=` params.
- **Transcript may not exist** — auto-generated captions exist for most popular videos but not all. A 404 means no captions for that language.
- **JSON is default** — `fmt=text` gives plain newline-separated lines, `fmt=srt` gives SRT subtitles with timestamps. No `fmt` flag gives JSON with `text`, `start`, `duration` fields.
- **Timestamps are seconds** — `start` and `duration` in JSON are float seconds, not milliseconds.
- **List languages first** — `GET /{video_id}/languages` returns `{"languages":[{"code":"en","name":"English"}]}`. Use it to pick a `lang` instead of guessing; a transcript 404 means that language is not available.
