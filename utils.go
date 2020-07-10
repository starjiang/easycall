package easycall

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/starjiang/elog"
)

const (
	EASYCALL_ETCD_SERVICE_PATH     = "/easycall/services"
	EASYCALL_ETCD_CONFIG_PATH      = "/easycall/configs"
	EASYCALL_CONFIG_PATH           = "./conf"
	EASYCALL_CONFIG_CHECK_INTERVAL = 60
	EASYCALL_WRITE_QUEUE_SIZE      = 100
	ETCD_KEEPLIVE_TIMEOUT          = 10
	ETCD_HEARTBEAT_INTEVAL         = 5
	ETCD_CONNECT_TIMEOUT           = 3
)

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetLocalIp() string {

	name, err := os.Hostname()
	if err != nil {
		elog.Errorf("get hostname fail: %v\n", err)
		return "127.0.0.1"
	}

	ipList, err := net.LookupHost(name)
	if err != nil {
		elog.Errorf("get lookup up addr: %v\n", err)
		return "127.0.0.1"
	}

	for _, ip := range ipList {
		if !strings.HasPrefix(ip, "127.") {
			return ip
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		elog.Errorf("get local ip fail. %s\n", err.Error())
		return "127.0.0.1"
	} else {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "127.0.0.1"
}

func GetTimeNow() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetTimeNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetTimeNowDate() string {
	return time.Now().Format("2006-01-02")
}

func FileIsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func PanicHandler() {
	if err := recover(); err != nil {
		elog.Error("Panic Exception:", err)
		elog.Error(string(debug.Stack()))
	}
}

func PanicHandlerExit() {
	if err := recover(); err != nil {
		elog.Error("Panic Exception:", err)
		elog.Error(string(debug.Stack()))
		elog.Error("************Program Exit************")
		os.Exit(0)
	}
}
