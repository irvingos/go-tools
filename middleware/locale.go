package middleware

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/irvingos/go-tools/i18n"
)

func Locale(defaultLocale i18n.Locale) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(consts.HTTP_ACCEPT_LANGUAGE)

		locale := defaultLocale
		locales := ParseAcceptLanguage(header)
		if len(locales) > 0 {
			locale = locales[0]
		}

		c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), locale))
		c.Next()
	}
}

type langQ struct {
	lang i18n.Locale
	q    float64 // 权重
}

func ParseAcceptLanguage(header string) []i18n.Locale {
	if header == "" {
		return nil
	}

	parts := strings.Split(header, ",")
	var langs []langQ

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		q := 1.0
		if strings.Contains(p, ";q=") {
			seg := strings.Split(p, ";q=")
			p = seg[0]
			if v, err := strconv.ParseFloat(seg[1], 64); err == nil {
				q = v
			}
		}

		langs = append(langs, langQ{lang: normalize(p), q: q})
	}

	sort.Slice(langs, func(i, j int) bool {
		return langs[i].q > langs[j].q
	})

	var res []i18n.Locale
	for _, l := range langs {
		res = append(res, l.lang)
	}

	return res
}

func normalize(lang string) i18n.Locale {
	lang = strings.TrimSpace(lang)

	switch {
	case strings.HasPrefix(lang, "zh"):
		return i18n.LocaleZH
	case strings.HasPrefix(lang, "en"):
		return i18n.LocaleEN
	default:
		return i18n.Locale(lang)
	}
}
