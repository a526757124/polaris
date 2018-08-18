package httpx

import (
	"testing"
	"fmt"
)

func TestHttpHead(t *testing.T) {
	fmt.Println(HttpHead("http://www.baidu.com/"))
}
