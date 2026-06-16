package redis

import (
	"context"
	"strings"
	"time"

	"github.com/irvingos/go-tools/logx"
	"github.com/irvingos/go-tools/notification"
	goredis "github.com/redis/go-redis/v9"
)

type DedupConfig struct {
	KeyPrefix string
	TTL       time.Duration
}

func defaultDedupConfig() DedupConfig {
	return DedupConfig{
		KeyPrefix: "notification:dedup:",
		TTL:       1 * time.Hour,
	}
}

type dedupNotifier struct {
	base   notification.Notifier
	client goredis.Cmdable
	cfg    DedupConfig
}

func NewDedupNotifier(base notification.Notifier, client goredis.Cmdable, cfg DedupConfig) notification.Notifier {
	return &dedupNotifier{
		base:   base,
		client: client,
		cfg:    normalizeDedupConfig(cfg),
	}
}

func (n *dedupNotifier) Channel() notification.Channel {
	return n.base.Channel()
}

func (n *dedupNotifier) Send(ctx context.Context, msg notification.Message) error {
	key := strings.TrimSpace(msg.Key)
	if key == "" {
		return n.base.Send(ctx, msg)
	}

	ok, err := n.client.SetNX(ctx, n.redisKey(key), "1", n.cfg.TTL).Result()
	if err != nil {
		logx.Errorf("notification dedup check failed key=%s: %v", key, err)
		return n.base.Send(ctx, msg)
	}
	if !ok {
		logx.Infof("notification deduplicated key=%s", key)
		return nil
	}

	return n.base.Send(ctx, msg)
}

func (n *dedupNotifier) redisKey(key string) string {
	return n.cfg.KeyPrefix + key
}

func normalizeDedupConfig(cfg DedupConfig) DedupConfig {
	defaults := defaultDedupConfig()
	if strings.TrimSpace(cfg.KeyPrefix) == "" {
		cfg.KeyPrefix = defaults.KeyPrefix
	}
	if cfg.TTL <= 0 {
		cfg.TTL = defaults.TTL
	}
	return cfg
}
