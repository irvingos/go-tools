package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	return &SSEWriter{
		w:       w,
		flusher: w.(http.Flusher),
	}
}

type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (s *SSEWriter) Event(event string, data any) error {
	if event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return err
		}
	}
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(s.w, "data: %s\n", b); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(s.w, "\n")
	if err == nil {
		s.flusher.Flush()
	}
	return err
}

func (s *SSEWriter) Done() {
	_ = s.Event("done", nil)
}
