package hystrix

import (
	"testing"
	"math/rand"
	"strconv"
	"fmt"
	"time"
)
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
func getRandNum() string{
	return strconv.Itoa(rnd.Intn(100))
}

func TestGetHystrix(t *testing.T) {
	key := getRandNum()
	fmt.Println(key, GetHystrix(key))
}

func BenchmarkGetHystrix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHystrix(getRandNum())
	}
}
