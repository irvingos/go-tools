package notification

import "context"

type Channel string

const (
	ChannelDingTalk Channel = "dingtalk"
	ChannelWebhook  Channel = "webhook"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Message struct {
	Key     string
	Title   string
	Content string
	Level   Level
	Targets []string
	Meta    map[string]any
}

type Notifier interface {
	Channel() Channel
	Send(ctx context.Context, msg Message) error
}
