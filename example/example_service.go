package main

import (
	"flag"

	"github.com/starjiang/easycall"
	"github.com/starjiang/elog"
)

type ProfileService struct {
}

func (ps *ProfileService) GetProfile(req *easycall.Request, resp *easycall.Response) {

	user := &UserInfo{}
	req.GetBody(user)
	//elog.Infof("head=%v,body=%v", req.GetHead(), user)

	respBody := make(map[string]interface{})
	respBody["name"] = "jiangyouxing"
	respBody["email"] = "starjiang@gmail.com"

	// index := rand.Intn(10)

	// time.Sleep(time.Millisecond * time.Duration(index))

	resp.SetBody(respBody)
}

func Middleware1(req *easycall.Request, resp *easycall.Response, client *easycall.EasyConnection, next *easycall.MiddlewareInfo) {
	user := &UserInfo{}
	req.GetBody(user)
	elog.Infof("head1=%v,body1=%v", req.GetHead(), user)
	next.Middleware(req, resp, client, next.Next)
}

func Middleware2(req *easycall.Request, resp *easycall.Response, client *easycall.EasyConnection, next *easycall.MiddlewareInfo) {
	user := &UserInfo{}
	req.GetBody(user)
	elog.Infof("head2=%v,body2=%v", req.GetHead(), user)

	// resp.SetHead(req.GetHead()).SetBody(make(map[string]interface{}))

	// respPkg := easycall.NewPackageWithBody(resp.GetFormat(), resp.GetHead(), resp.GetBody())

	// pkgData, err := respPkg.EncodeWithBody()
	// if err != nil {
	// 	elog.Error("encode pkg fail:", err)
	// }
	// client.Send(pkgData)
	next.Middleware(req, resp, client, next.Next)
}

type UserInfo struct {
	Name string `json:"name"`
	Uid  uint64 `json:"uid"`
	Seq  uint64 `json:"seq"`
}

var port int

func init() {
	flag.IntVar(&port, "port", 8001, "listen port")
}

type ApmReport struct {
	service string
}

func (ar *ApmReport) OnData(data map[string]*easycall.ApmMonitorStatus) {
	elog.Error(data["GetProfile"])
}

func main() {
	flag.Parse()
	defer elog.Flush()
	context := easycall.NewServiceContext([]string{"127.0.0.1:2379"})
	context.CreateService("profile", port, &ProfileService{}, 100)
	//context.CreateService("profile1", port+1, &ProfileService{}, 100)
	//context.AddMiddleware("profile", Middleware1)
	//context.AddMiddleware("profile", Middleware2)
	context.AddMiddleware("profile", easycall.NewApmMonitor(&ApmReport{"profile"}).Middleware)
	context.StartAndWait()
}
