package resp

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJson(t *testing.T) {
	res := &Response{}
	raw, _ := json.Marshal(res)
	fmt.Println(string(raw))
}
