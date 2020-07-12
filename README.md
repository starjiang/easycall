# easycall微服务框架

easycall 是一款go 微服务框架，轻量,高性能，类似dubbo,motan 微服务框架，主要特性如下：
========================
* 轻量，依赖少，代码量少，方便阅读，虽然轻量，功能齐全
* 大量使用goroutine 池，连接池，提升请求处理速度
* 完全 scheme free 接口调用,无需定义interface 接口文件
* 支持跨语言调用python,php,java,c/c++等，凡是支持json/msgpack 序列化的语言都没问题
* 数据序列化支持 json/msgpack
* 客户端支持同步，异步调用
* 负载均衡支持随机，轮询，随机权重，动态负载，hash 五种负载均衡算法
* 已经集成配置中心,实现配置动态加载，集中管理
* 支持熔断机制,方便服务降级
* 支持中间件处理机制，方便扩展（比如性能统计，登录校验等等）
* 支持API网关，网关支持http json,easycall协议

easycall 性能
========================
* 在 2.5 GHz Intel Core i7 4核8线程(macbook pro 2015 版) 机器上可以压测到7w qps/s,压测程序跟被压测程序部署在同一台机器上，如果分开部署，估计qps 会更高

微服务例子
=====================================
```
package main

import (
	"flag"

	"github.com/starjiang/easycall"
	"github.com/starjiang/elog"
)

//微服务申明
type ProfileService struct {
}

//微服务方法实现,req 为请求封装，resp 为返回封装
func (ps *ProfileService) GetProfile(req *easycall.Request, resp *easycall.Response) {

	user := &UserInfo{}
	req.GetBody(user) //读取请求body
	elog.Infof("head=%v,body=%v", req.GetHead(), user)

	respBody := make(map[string]interface{})
	respBody["name"] = "jiangyouxing"
	respBody["email"] = "starjiang@gmail.com"

	resp.SetHead(req.GetHead()).SetBody(respBody) //设置返回head,body,返回结果
}
//调用请求body
type UserInfo struct {
	Name string `json:"name"`
	Uid  uint64 `json:"uid"`
	Seq  uint64 `json:"seq"`
}

var port int

func init() {
	flag.IntVar(&port, "port", 8001, "listen port")
}

func main() {
	flag.Parse()
	defer elog.Flush() //log
	context := easycall.NewServiceContext([]string{"127.0.0.1:2379"}) //创建微服务上下文，用于管理微服务，参数为etcd endpoints 列表
	context.CreateService("profile", port, &ProfileService{}, 100) //创建微服务profile,端口为port,微服务实现为ProfileService,权重为100
	context.StartAndWait()//启动微服务，并等待
}
```
调用列子
===============
```
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

	easyClient := easycall.NewServiceClient([]string{"127.0.0.1:2379"}, "profile", 100, easycall.LB_ACTIVE)//创建微服务调用客户端，参数为etcd endpoints,要调用的微服务名,100 是连接池大小,负载均衡 是easycall.LB_ACTIVE（动态请求负载均衡） 

	reqBody := make(map[string]interface{}) //请求body
	respBody := make(map[string]interface{}) //返回body

	//请求profile微服务的 GetProfile 接口，time.Second 为超时时间，设定为1秒
	err := easyClient.Request("GetProfile", reqBody, &respBody, time.Second)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp=", respBody) //打印返回body
}

```
