package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/starjiang/easycall"
	"github.com/starjiang/elog"
)

func main() {

	flag.Parse()
	defer elog.Flush()
	easyClient := easycall.NewServiceClient([]string{"127.0.0.1:2379"}, "profile", 100, easycall.LB_ACTIVE)

	for i := 0; i < 1000; i++ {
		go func() {
			reqBody := make(map[string]interface{})
			respBody := make(map[string]interface{})
			err := easyClient.Request("GetProfile", reqBody, &respBody, time.Second)
			if err != nil {

				fmt.Println(err)
				return
			}
			fmt.Println("resp=", respBody)

		}()
		time.Sleep(time.Second * 1)
	}
	time.Sleep(time.Second * 4)
}
