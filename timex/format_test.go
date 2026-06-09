package timex

import (
	"fmt"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	fmt.Println(Second.Format(time.Now()))
}
