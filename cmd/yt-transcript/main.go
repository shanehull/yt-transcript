// Command yt-transcript fetches YouTube transcripts from the command line.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	yt "github.com/shanehull/yt-transcript"
)

func main() {
	lang := flag.String("lang", "en", "language code for transcript")
	format := flag.String("fmt", "text", "output format: text, json, srt")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <video_id>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -lang en dQw4w9WgXcQ\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	videoID := flag.Arg(0)

	client := yt.NewClient()
	segments, err := client.FetchTranscript(context.Background(), videoID, *lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch *format {
	case "json":
		outputJSON(segments)
	case "srt":
		outputSRT(segments)
	case "text":
		outputText(segments)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format %q (use text, json, or srt)\n", *format)
		os.Exit(1)
	}
}

func outputText(segments []yt.Segment) {
	for _, s := range segments {
		fmt.Println(s.Text)
	}
}

func outputJSON(segments []yt.Segment) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(segments)
}

func outputSRT(segments []yt.Segment) {
	for i, s := range segments {
		fmt.Printf("%d\n", i+1)
		fmt.Printf("%s --> %s\n", srtTime(s.Start), srtTime(s.Start+s.Duration))
		fmt.Printf("%s\n\n", strings.TrimSpace(s.Text))
	}
}

func srtTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
