package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/starjiang/easycall-go"
)

var requestCount = 0

func main() {

	flag.Parse()

	if len(os.Args) < 3 {
		fmt.Println("usage:", os.Args[0]+" concurrent num")
		return
	}

	c1, _ := strconv.Atoi(os.Args[1])
	n1, _ := strconv.Atoi(os.Args[2])

	var c int32 = int32(c1)
	var n int32 = int32(n1)

	var pc int32 = 0
	easyClient := easycall.NewServiceClient([]string{"127.0.0.1:2379"}, "profile", c1, easycall.LB_ROUND_ROBIN)

	t1 := time.Now()

	for i := 0; i < int(c); i++ {
		requestCount++
		go func() {
			for {
				respBody := make(map[string]interface{})
				err := easyClient.Request("GetProfile", nil, respBody, time.Second*1)
				if err != nil {
					fmt.Println(err)
				}
				atomic.AddInt32(&pc, 1)

				if pc >= n {
					break
				}
			}
		}()
	}

	for {
		if pc >= n {
			break
		}
		time.Sleep(time.Millisecond * 1)
	}

	t2 := time.Now()

	duration := t2.Sub(t1)

	fmt.Println("spend", duration, ",qps", float64(n)/duration.Seconds())

	fmt.Println("completed")
}
