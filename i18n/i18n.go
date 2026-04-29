package i18n

import "context"

type Locale string

const (
	LocaleEN Locale = "en-US"
	LocaleZH Locale = "zh-CN"
)

func (l Locale) String() string {
	return string(l)
}

// key for context
type localeKey struct{}

func WithLocale(ctx context.Context, locale Locale) context.Context {
	return context.WithValue(ctx, localeKey{}, locale)
}

func LocaleFrom(ctx context.Context) Locale {
	locale, ok := ctx.Value(localeKey{}).(Locale)
	if !ok {
		return LocaleEN
	}
	return locale
}
