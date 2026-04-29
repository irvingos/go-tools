package resp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/irvingos/go-tools/errorx"
	"github.com/irvingos/go-tools/i18n"
)

func TestJson(t *testing.T) {
	res := &Response{}
	raw, _ := json.Marshal(res)
	fmt.Println(string(raw))
}

func TestResponderErrorWithFormattedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "zh-CN.yaml"), []byte("common.bad_request: field %s is required\n"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	catalog, err := i18n.LoadDir(dir, i18n.LocaleZH)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)
	g.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	NewResponder(catalog).Error(g, errorx.Format(errorx.ErrBadRequest, "name"))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got, want := resp.Code, errorx.ErrBadRequest.Code(); got != want {
		t.Fatalf("code = %d, want %d", got, want)
	}
	if got, want := resp.Message, "field name is required"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
	if got, want := resp.Detail, ""; got != want {
		t.Fatalf("detail = %q, want %q", got, want)
	}
}
