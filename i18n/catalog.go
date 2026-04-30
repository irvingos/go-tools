package i18n

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/irvingos/go-tools/errorx"
)

type Catalog struct {
	defaultLocale Locale
	messages      map[Locale]map[string]string
}

func (c *Catalog) Message(ctx context.Context, key string) string {
	locale := LocaleFrom(ctx)
	if m := c.messages[locale]; m != nil {
		if msg, ok := m[key]; ok {
			return msg
		}
	}
	if m := c.messages[c.defaultLocale]; m != nil {
		if msg, ok := m[key]; ok {
			return msg
		}
	}

	return key
}

func (c *Catalog) ErrorMessage(ctx context.Context, err error) string {
	var errorCode errorx.Error
	if ok := errors.As(err, &errorCode); ok {
		msg := c.Message(ctx, errorCode.I18nKey())
		var argError errorx.FormattedError
		if errors.As(err, &argError) {
			return fmt.Sprintf(msg, argError.Args()...)
		}
		return msg
	}
	return c.ErrorMessage(ctx, errorx.ErrInternalServerError)
}

func LoadDir(dir string, defaultLocale Locale) (*Catalog, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	messages := make(map[Locale]map[string]string)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		locale := Locale(strings.TrimSuffix(file.Name(), ext))

		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		msg, err := MessagesFromYAML(content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.Name(), err)
		}

		messages[locale] = msg
	}

	return &Catalog{
		defaultLocale: defaultLocale,
		messages:      messages,
	}, nil
}

// MessagesFromYAML 将 YAML 解析为扁平 map[string]string。
// 支持两种写法（可混用）：
//   - 嵌套：common: { bad_request: "..." }
//   - 扁平：common.bad_request: "..."
//
// 键冲突（解析后得到同一完整键）时返回错误。
func MessagesFromYAML(content []byte) (map[string]string, error) {
	if len(bytes.TrimSpace(content)) == 0 {
		return map[string]string{}, nil
	}
	var root map[string]interface{}
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, err
	}
	if root == nil {
		return map[string]string{}, nil
	}
	return flattenMessageRoot(root)
}

func flattenMessageRoot(root map[string]interface{}) (map[string]string, error) {
	out := make(map[string]string)
	for k, v := range root {
		if err := walkFlatten("", k, v, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func walkFlatten(prefix, key string, v interface{}, out map[string]string) error {
	full := key
	if prefix != "" {
		full = prefix + "." + key
	}

	switch t := v.(type) {
	case string:
		if _, exists := out[full]; exists {
			return fmt.Errorf("i18n: duplicate key %q", full)
		}
		out[full] = t
	case map[string]interface{}:
		for nk, nv := range t {
			if err := walkFlatten(full, nk, nv, out); err != nil {
				return err
			}
		}
	case map[interface{}]interface{}:
		for nk, nv := range t {
			ks, ok := nk.(string)
			if !ok {
				return fmt.Errorf("i18n: non-string map key under %q", full)
			}
			if err := walkFlatten(full, ks, nv, out); err != nil {
				return err
			}
		}
	case nil:
		return fmt.Errorf("i18n: nil value for key %q", full)
	default:
		return fmt.Errorf("i18n: invalid value type for key %q: want string or map, got %T", full, v)
	}
	return nil
}
