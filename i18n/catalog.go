package i18n

import (
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

		msg := make(map[string]string)
		err = yaml.Unmarshal(content, &msg)
		if err != nil {
			return nil, err
		}

		messages[locale] = msg
	}

	return &Catalog{
		defaultLocale: defaultLocale,
		messages:      messages,
	}, nil
}
