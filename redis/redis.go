package redis

import (
	"context"
	"fmt"
	"strings"

	redis "github.com/redis/go-redis/v9"
)

type Config struct {
	Mode     string
	Addr     string
	Password string
}

func (c Config) GetAddrs() []string {
	raw := strings.Split(c.Addr, ",")
	addrs := make([]string, 0, len(raw))
	for _, addr := range raw {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

type Client interface {
	redis.Cmdable
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	PSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	Close() error
}

func NewClient(ctx context.Context, c Config) (Client, error) {
	var client Client

	switch c.Mode {
	case "single":
		client = redis.NewClient(&redis.Options{
			Addr:     c.Addr,
			Password: c.Password,
		})
	case "cluster":
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    c.GetAddrs(),
			Password: c.Password,
		})
	default:
		return nil, fmt.Errorf("unknown mode %s", c.Mode)
	}

	if _, err := client.Ping(ctx).Result(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
