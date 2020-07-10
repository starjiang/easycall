# easycall-go easycall微服务框架go语言版本
code for service
=====================================
```
type ProfileService struct {
}

func (ps *ProfileService) GetProfile(req *easycall.Request, resp *easycall.Response) {
	user := &UserInfo{}
	req.GetBody(user)
	elog.Infof("head=%v,body=%v", req.GetHead(), user)
	resp.SetHead(req.GetHead()).SetBody(make(map[string]interface{}))
}

type UserInfo struct {
	Name string `json:"name"`
	Uid  uint64 `json:"uid"`
	Seq  uint64 `json:"seq"`
}

func main() {
	flag.Parse()
	defer elog.Flush()

	service := easycall.NewEasyService([]string{"172.28.2.162:2181"})
	service.CreateService("profile", 8003, &ProfileService{}, 100)
	service.StartAndWait()
}
```
code for client
===============
```
func main() {

	flag.Parse()
	defer elog.Flush()

	easyClient := easycall.NewEasyClient([]string{"172.28.2.162:2181"}, 10, easycall.LB_ACTIVE)

	for i := 0; i < 100; i++ {
		go func() {
			head := easycall.NewEasyHead().SetService("profile").SetMethod("GetProfile")
			body := make(map[string]interface{})
			respPkg, err := easyClient.Request(easycall.FORMAT_MSGPACK, head, body, time.Second*3)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("resp=", respPkg.GetHead())
			}
		}()
		time.Sleep(time.Second * 1)
	}
	time.Sleep(time.Second * 1)
}
```
