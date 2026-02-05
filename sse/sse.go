package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/irvingos/go-tools/safego"
)

func NewSSEWriter(ctx context.Context, w http.ResponseWriter, heartbeatInterval time.Duration) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not implement http.Flusher")
	}

	writer := &SSEWriter{
		ctx:               ctx,
		w:                 w,
		flusher:           flusher,
		heartbeatInterval: heartbeatInterval,
		lastSendAt:        time.Now(),
		done:              make(chan struct{}),
		closed:            false,
	}
	if heartbeatInterval > 0 {
		safego.Go(ctx, func(ctx context.Context) {
			writer.startHeartbeat()
		})
	}
	return writer, nil
}

type SSEWriter struct {
	ctx               context.Context
	w                 http.ResponseWriter
	flusher           http.Flusher
	heartbeatInterval time.Duration

	lastSendAt time.Time
	mu         sync.Mutex
	done       chan struct{}
	closed     bool
}

func (s *SSEWriter) startHeartbeat() {
	for {
		s.mu.Lock()
		next := s.lastSendAt.Add(s.heartbeatInterval)
		s.mu.Unlock()

		d := max(time.Until(next), 0)

		timer := time.NewTimer(d)

		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case <-s.done:
			timer.Stop()
			return
		case <-timer.C:
			// 再次确认这段时间内是否有人发送过数据
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				return
			}
			should := !time.Now().Before(s.lastSendAt.Add(s.heartbeatInterval))
			s.mu.Unlock()

			if should {
				_ = s.Heartbeat()
			}
		}
	}
}

func (s *SSEWriter) Event(event string, data any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("SSEWriter is closed")
	}

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
	if _, err := fmt.Fprint(s.w, "\n"); err != nil {
		return err
	}

	s.flusher.Flush()
	s.lastSendAt = time.Now()
	return nil
}

func (s *SSEWriter) Comment(text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("SSEWriter is closed")
	}

	if _, err := fmt.Fprintf(s.w, ": %s\n\n", text); err != nil {
		return err
	}
	s.flusher.Flush()
	s.lastSendAt = time.Now()
	return nil
}

func (s *SSEWriter) Done() error {
	// 先發送 done 事件
	err := s.Event("done", nil)

	// 然後標記為關閉並停止 heartbeat
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.done)
	}
	s.mu.Unlock()

	return err
}

func (s *SSEWriter) Heartbeat() error {
	return s.Comment("ping")
}
