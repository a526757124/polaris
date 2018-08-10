package leakybucket

import (
	"fmt"
	"time"
	"testing"
)

func Test_Run(t *testing.T){
	lb := newLeakyBucket()
	fmt.Println(time.Now())
	for i:=0;i<1000;i++{
		go func(){
			if lb.acquire(){
				//
			}else{
				fmt.Println(i, "limit!")
			}
		}()
	}
	fmt.Println(time.Now())
}
