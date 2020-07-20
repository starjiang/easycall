package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/starjiang/easycall"
	"github.com/starjiang/elog"
)

var port int
var httpport int
var endpoints string
var poolSize int
var timeout int
var easyClient *easycall.EasyClient

func init() {
	flag.IntVar(&port, "port", 8088, "listen port")
	flag.IntVar(&httpport, "httpport", 8080, "http listen port")
	flag.StringVar(&endpoints, "endpoints", "127.0.0.1:2379", "etcd endpoints")
	flag.IntVar(&poolSize, "poolsize", 100, "pool size")
	flag.IntVar(&timeout, "timeout", 5, "pool size")
}

func ProxyMiddleware(reqPkg *easycall.EasyPackage, client *easycall.EasyConnection, next *easycall.GWMiddlewareInfo) {
	respPkg, err := easyClient.RequestWithHead(reqPkg.GetFormat(), reqPkg.GetHead(), reqPkg.GetBodyData(), time.Duration(timeout)*time.Second)

	if err != nil {
		respPkg = easycall.NewPackageWithBody(reqPkg.GetFormat(), reqPkg.GetHead(), nil)
		systemError, ok := err.(*easycall.SystemError)
		if ok {
			respPkg.GetHead().SetRet(systemError.GetRet()).SetMsg(systemError.GetMsg())
		} else {
			respPkg.GetHead().SetRet(easycall.ERROR_INTERNAL_ERROR).SetMsg(err.Error())
		}
		respData, err := respPkg.EncodeWithBody()
		if err != nil {
			elog.Error(err)
			return
		}
		client.Send(respData)
	}
	client.Send(respPkg.GetPkgData())

	if next != nil {
		next.Middleware(reqPkg, client, next.Next)
	}
}

func HttpProxyMiddleware(w http.ResponseWriter, r *http.Request, next *easycall.HttpMiddlewareInfo) {

	head := easycall.NewEasyHead()

	service := ""
	method := ""

	paths := strings.Split(r.URL.Path, "/")

	if len(paths) == 2 {
		service = r.Header.Get("X-Easycall-Service")
		method = r.Header.Get("X-Easycall-Method")
	} else if len(paths) == 3 {
		service = paths[1]
		method = paths[2]
	}

	if service == "" || method == "" {
		http.Error(w, "service or method is empty", http.StatusBadRequest)
		return
	}

	head.SetService(service)
	head.SetMethod(method)

	if r.Header.Get("X-Easycall-Uid") != "" {
		uid, _ := strconv.ParseUint(r.Header.Get("X-Easycall-Uid"), 10, 0)
		head.SetUid(uid)
	}

	if r.Header.Get("X-Easycall-Token") != "" {
		head.SetToken(r.Header.Get("X-Easycall-Token"))
	}

	head.SetRequestIp(r.RemoteAddr)

	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respPkg, err := easyClient.RequestWithHead(easycall.FORMAT_JSON, head, bodyData, time.Duration(timeout)*time.Second)

	if err != nil {
		systemError, ok := err.(*easycall.SystemError)
		if ok {
			w.Header().Set("X-Easycall-Ret", strconv.Itoa(systemError.GetRet()))
			http.Error(w, systemError.GetMsg(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if respPkg.GetHead().GetRet() != 0 {
		if respPkg.GetHead().GetRet() < easycall.ERROR_MAX_SYSTEM_CODE {
			w.Header().Set("X-Easycall-Ret", strconv.Itoa(respPkg.GetHead().GetRet()))
			http.Error(w, respPkg.GetHead().GetMsg(), http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("X-Easycall-Ret", strconv.Itoa(respPkg.GetHead().GetRet()))
			http.Error(w, respPkg.GetHead().GetMsg(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(respPkg.GetBodyData())

	if next != nil {
		next.Middleware(w, r, next.Next)
	}
}

func main() {
	flag.Parse()
	defer elog.Flush()

	easyClient = easycall.NewEasyClient(strings.Split(endpoints, ","), poolSize, easycall.LB_ACTIVE)

	gwhandler := easycall.NewGatewayHandler(nil)
	gwhandler.AddMiddleware(ProxyMiddleware)

	httphandler := easycall.NewHttpHandler(nil)
	httphandler.AddMiddleware(HttpProxyMiddleware)
	gateway := &easycall.Gateway{}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		elog.Infof("start gateway server,listen port:%v", port)
		err := gateway.StartGWServer(port, gwhandler)
		if err != nil {
			elog.Error(err)
		}
		wg.Done()
	}()

	go func() {
		elog.Infof("start http gateway server,listen port:%v", httpport)
		err := gateway.StartHttpGWServer(httpport, httphandler)
		if err != nil {
			elog.Error(err)
		}
		wg.Done()
	}()

	wg.Wait()

}
