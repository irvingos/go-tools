package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/irvingos/go-tools/logx"
	"github.com/irvingos/go-tools/notification"
	notificationredis "github.com/irvingos/go-tools/notification/redis"
	"github.com/irvingos/go-tools/timex"
	goredis "github.com/redis/go-redis/v9"
)

func TestMain(m *testing.M) {
	logx.Init(&logx.Options{
		Format:          logx.Format("json"),
		TimestampFormat: timex.Second,
		Level:           logx.Level(4),
	})
	os.Exit(m.Run())
}

type stubNotifier struct {
	channel notification.Channel
	calls   int
}

func (s *stubNotifier) Channel() notification.Channel {
	return s.channel
}

func (s *stubNotifier) Send(_ context.Context, _ notification.Message) error {
	s.calls++
	return nil
}

func newDedupNotifier(t *testing.T, server *miniredis.Miniredis, base notification.Notifier) notification.Notifier {
	t.Helper()

	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return notificationredis.NewDedupNotifier(base, client, notificationredis.DedupConfig{
		KeyPrefix: "notification:dedup:",
		TTL:       time.Minute,
	})
}

func TestDedupNotifier_SendAllowsDuplicateWhenKeyIsEmpty(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("run miniredis: %v", err)
	}
	defer server.Close()

	base := &stubNotifier{channel: notification.ChannelDingTalk}
	notifier := newDedupNotifier(t, server, base)

	msg := notification.Message{Title: "no key"}
	if err := notifier.Send(context.Background(), msg); err != nil {
		t.Fatalf("first send: %v", err)
	}
	if err := notifier.Send(context.Background(), msg); err != nil {
		t.Fatalf("second send: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("calls=%d", base.calls)
	}
}

func TestDedupNotifier_SendDeduplicatesByMessageKey(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("run miniredis: %v", err)
	}
	defer server.Close()

	base := &stubNotifier{channel: notification.ChannelDingTalk}
	notifier := newDedupNotifier(t, server, base)

	msg := notification.Message{Key: "device-1|1001", Title: "device offline"}
	if err := notifier.Send(context.Background(), msg); err != nil {
		t.Fatalf("first send: %v", err)
	}
	if err := notifier.Send(context.Background(), msg); err != nil {
		t.Fatalf("second send: %v", err)
	}
	if base.calls != 1 {
		t.Fatalf("calls=%d", base.calls)
	}
}

func TestDedupNotifier_SendUsesDifferentMessageKeysIndependently(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("run miniredis: %v", err)
	}
	defer server.Close()

	base := &stubNotifier{channel: notification.ChannelDingTalk}
	notifier := newDedupNotifier(t, server, base)

	first := notification.Message{Key: "device-1|1001", Title: "device offline"}
	second := notification.Message{Key: "device-2|1001", Title: "device offline"}
	if err := notifier.Send(context.Background(), first); err != nil {
		t.Fatalf("first send: %v", err)
	}
	if err := notifier.Send(context.Background(), second); err != nil {
		t.Fatalf("second send: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("calls=%d", base.calls)
	}
}

func TestDedupNotifier_SendSharesDedupStateAcrossInstances(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("run miniredis: %v", err)
	}
	defer server.Close()

	baseA := &stubNotifier{channel: notification.ChannelDingTalk}
	baseB := &stubNotifier{channel: notification.ChannelDingTalk}
	notifierA := newDedupNotifier(t, server, baseA)
	notifierB := newDedupNotifier(t, server, baseB)

	msg := notification.Message{Key: "device-1|1001", Title: "device offline"}
	if err := notifierA.Send(context.Background(), msg); err != nil {
		t.Fatalf("first send: %v", err)
	}
	if err := notifierB.Send(context.Background(), msg); err != nil {
		t.Fatalf("second send: %v", err)
	}
	if baseA.calls != 1 {
		t.Fatalf("baseA calls=%d", baseA.calls)
	}
	if baseB.calls != 0 {
		t.Fatalf("baseB calls=%d", baseB.calls)
	}
}

func TestDedupNotifier_SendAllowsDuplicateAfterTTLExpires(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("run miniredis: %v", err)
	}
	defer server.Close()

	base := &stubNotifier{channel: notification.ChannelDingTalk}
	notifier := newDedupNotifier(t, server, base)

	msg := notification.Message{Key: "device-1|1001", Title: "device offline"}
	if err := notifier.Send(context.Background(), msg); err != nil {
		t.Fatalf("first send: %v", err)
	}

	server.FastForward(2 * time.Minute)

	if err := notifier.Send(context.Background(), msg); err != nil {
		t.Fatalf("second send: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("calls=%d", base.calls)
	}
}
