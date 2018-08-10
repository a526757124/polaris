package tokenbucket

import (
	"fmt"
	"time"
	"testing"
)

func Test_Run(t *testing.T){
	tb := newTokenBucket()
	fmt.Println(time.Now())
	for i:=0;i<1000;i++{
		go func(){
			if tb.acquire(1){
				//
			}else{
				fmt.Println(i, "limit!")
			}
		}()
	}
	fmt.Println(time.Now())
}
