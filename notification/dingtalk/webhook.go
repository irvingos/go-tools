package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/irvingos/go-tools/logx"
	"github.com/irvingos/go-tools/notification"
)

var (
	ErrWebhookURLRequired = errors.New("dingtalk webhook: webhook url is required")
	ErrSendFailed         = errors.New("dingtalk webhook: send failed")
)

type Config struct {
	WebhookURL     string
	Secret         string
	Timeout        time.Duration
	DefaultTargets string
}

type webhookNotifier struct {
	webhookURL     string
	secret         string
	timeout        time.Duration
	defaultTargets []string

	httpClient *http.Client
}

func NewWebhookNotifier(cfg Config) (notification.Notifier, error) {
	webhookURL := strings.TrimSpace(cfg.WebhookURL)
	if webhookURL == "" {
		return nil, ErrWebhookURLRequired
	}
	return &webhookNotifier{
		webhookURL:     webhookURL,
		secret:         cfg.Secret,
		timeout:        cfg.timeout(),
		defaultTargets: strings.Split(cfg.DefaultTargets, ","),
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: cfg.timeout(),
		},
	}, nil
}

func (n *webhookNotifier) Channel() notification.Channel {
	return notification.ChannelDingTalk
}

func (n *webhookNotifier) Send(ctx context.Context, msg notification.Message) error {
	payload, err := json.Marshal(n.buildWebhookRequest(msg))
	if err != nil {
		return fmt.Errorf("dingtalk webhook: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookEndpoint(), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("dingtalk webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk webhook: send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("dingtalk webhook: read response: %w", err)
	}

	var resp webhookResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("dingtalk webhook: parse response body=%s: %w", string(body), err)
	}
	if resp.ErrCode != 0 {
		logx.Errorf("dingtalk webhook send failed title=%s response=%s", msg.Title, string(body))
		return fmt.Errorf("%w: errcode=%d errmsg=%s", ErrSendFailed, resp.ErrCode, resp.ErrMsg)
	}

	logx.Infof("dingtalk webhook sent title=%s level=%s", msg.Title, msg.Level)
	return nil
}

type webhookRequest struct {
	MsgType  string       `json:"msgtype"`
	Markdown markdownBody `json:"markdown"`
	At       *atBody      `json:"at,omitempty"`
}

type markdownBody struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type atBody struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

type webhookResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (c Config) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 10 * time.Second
}

func (n *webhookNotifier) webhookEndpoint() string {
	endpoint := n.webhookURL
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	query := url.Values{}
	for key, values := range u.Query() {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	if n.secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		query.Set("timestamp", timestamp)
		query.Set("sign", sign(n.secret, timestamp))
	}
	u.RawQuery = query.Encode()
	return u.String()
}

func sign(secret, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + secret))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (n *webhookNotifier) buildWebhookRequest(msg notification.Message) webhookRequest {
	title := strings.TrimSpace(msg.Title)
	targets := mergeTargets(msg.Targets, n.defaultTargets)
	text := formatMarkdownText(msg)
	at := buildAtBody(targets)

	req := webhookRequest{
		MsgType: "markdown",
		Markdown: markdownBody{
			Title: title,
			Text:  text,
		},
	}
	if at != nil {
		req.At = at
	}

	return req
}

func mergeTargets(targets []string, defaultTargets []string) []string {
	merged := make([]string, 0, len(targets)+len(defaultTargets))
	seen := make(map[string]struct{}, len(targets)+len(defaultTargets))
	for _, target := range append(targets, defaultTargets...) {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		key := strings.ToLower(target)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, target)
	}
	return merged
}

func formatMarkdownText(msg notification.Message) string {
	var b strings.Builder
	if msg.Level != "" {
		fmt.Fprintf(&b, "**[%s]** ", strings.ToUpper(string(msg.Level)))
	}
	if title := strings.TrimSpace(msg.Title); title != "" {
		b.WriteString("### ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if content := strings.TrimSpace(msg.Content); content != "" {
		b.WriteString(content)
	}
	if atText := formatAtText(msg.Targets); atText != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(atText)
	}
	return strings.TrimSpace(b.String())
}

func formatAtText(targets []string) string {
	var mobiles []string
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" || strings.EqualFold(target, "all") || target == "*" {
			continue
		}
		mobiles = append(mobiles, "@"+target)
	}
	return strings.Join(mobiles, " ")
}

func buildAtBody(targets []string) *atBody {
	if len(targets) == 0 {
		return nil
	}

	at := &atBody{}
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		if strings.EqualFold(target, "all") || target == "*" {
			at.IsAtAll = true
			return at
		}
		at.AtMobiles = append(at.AtMobiles, target)
	}
	if len(at.AtMobiles) == 0 && !at.IsAtAll {
		return nil
	}
	return at
}
