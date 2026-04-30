package i18n

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/irvingos/go-tools/errorx"
)

func TestCatalogErrorMessageWithArgs(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"en-US.yaml": "common.bad_request: field %s is required\n",
		"zh-CN.yaml": "common.bad_request: 字段 %s 必填\n",
	}
	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("WriteFile(%s): %v", name, err)
		}
	}

	catalog, err := LoadDir(dir, LocaleEN)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	err = errorx.Format(errorx.ErrBadRequest, "name")
	ctx := context.Background()
	ctx = WithLocale(ctx, LocaleEN)
	if got, want := catalog.ErrorMessage(ctx, err), "field name is required"; got != want {
		t.Fatalf("en message = %q, want %q", got, want)
	}
	ctx = WithLocale(ctx, LocaleZH)
	if got, want := catalog.ErrorMessage(ctx, err), "字段 name 必填"; got != want {
		t.Fatalf("zh message = %q, want %q", got, want)
	}
}
