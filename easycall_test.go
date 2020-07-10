package easycall

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"testing"

	"github.com/starjiang/elog"
	"github.com/vmihailenco/msgpack"
)

func TestBinary(t *testing.T) {

	var data = make([]byte, 100)
	buffer := bytes.NewBuffer(data)

	var a int16 = 10
	var b uint32 = 200
	var c uint64 = 23
	var d int16 = 10

	err := binary.Write(buffer, binary.BigEndian, a)
	if err != nil {
		fmt.Println(err)
	}
	binary.Write(buffer, binary.BigEndian, b)
	binary.Write(buffer, binary.BigEndian, c)
	binary.Write(buffer, binary.BigEndian, d)

	var a1 int8 = 20

	err1 := binary.Read(buffer, binary.BigEndian, &a1)
	if err1 != nil {
		fmt.Println(err1)
	}

	fmt.Printf("%v,%v\n", data, a1)
}

func TestReflect(t *testing.T) {

	pkg := &EasyPackage{}
	info := reflect.ValueOf(pkg)

	m := info.MethodByName("SetHead")

	head := &EasyHead{}
	head.Service = "profile"

	args := []reflect.Value{
		reflect.ValueOf(head),
	}

	m.Call(args)

	fmt.Println(pkg.GetHead().Service)

}
func TestJsonTag(t *testing.T) {
	head := &EasyHead{Service: "abc"}
	var buf bytes.Buffer
	encode := msgpack.NewEncoder(&buf).UseJSONTag(true)
	err := encode.Encode(head)
	if err != nil {
		fmt.Println(err)
		return
	}
	head.Service = ""
	fmt.Println(string(buf.Bytes()))
	decode := msgpack.NewDecoder(bytes.NewReader(buf.Bytes())).UseJSONTag(true)
	decode.Decode(head)
	fmt.Println(head)

}

func TestGetLocalIp(t *testing.T) {
	fmt.Println(GetLocalIp())
}

func TestRandIntn(t *testing.T) {

	for i := 0; i < 100; i++ {
		fmt.Println(rand.Intn(10))
	}

}

func TestFuncCaller(t *testing.T) {
	_, file, line, ok := runtime.Caller(0)
	fmt.Println(file, line, ok)
}

func TestEasyLog(t *testing.T) {
	defer elog.Flush()
	flag.Parse()
	elog.Info(111, "sdsdsdsd")
}

func TestJsonByteBuffer(t *testing.T) {
	head := NewEasyHead().SetService("profile").SetMethod("getProfile")

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(head)
	json.NewEncoder(&buf).Encode(head)
	fmt.Println(string(buf.Bytes()))
}

func TestFloat(t *testing.T) {

	f := float32(1) / float32(3)
	fmt.Println(f)
}

var count = 0

func getUserName() error {
	if count > 15 {
		return nil
	}
	count++
	return errors.New("invalid params")
}

func fail() error {
	fmt.Println("fall")
	return nil
}

func TestCircuitBreaker(t *testing.T) {

	for i := 0; i < 100; i++ {
		err := CbCall("hello", getUserName, nil)
		fmt.Println(err)
	}
}

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

func TestEasyConfig(t *testing.T) {

	config := NewEasyConfig()
	err := config.Load("./config.ini")
	if err != nil {
		fmt.Println(err)
	}

	tconfig := &TConfig{}
	fmt.Println(config.GetInt64("easycall.port", 10))
	config.GetConfig("easycall.", tconfig)
	fmt.Println(tconfig)

	tconfig1 := &TConfig{}
	config = GetConfig()
	fmt.Println(config.GetInt64("easycall.port", 10))
	config.GetConfig("easycall.", tconfig1)
	fmt.Println(tconfig1)

}

func TestSyncMap(t *testing.T) {

}
