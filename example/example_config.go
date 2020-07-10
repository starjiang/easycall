package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/starjiang/easycall"
)

type TConfig struct {
	ThreadNum int
	Name      string
	Port      int
	IsExsit   bool
	FailRate  float32
	UserList  []int64
	UserMap   map[string]int64
	UserMap1  map[string]bool
	UserMap2  map[string]float64
	UserMap3  map[string]string
}

func init() {

}

func main() {

	flag.Parse()

	config := easycall.NewRemoteEasyConfig([]string{"127.0.0.1:2379"}, "profile")
	tconfig := &TConfig{}
	fmt.Println(config.GetInt64("service.port", 10))
	config.GetConfig("easycall.", tconfig)
	fmt.Println(config)
	time.Sleep(time.Hour)
}
