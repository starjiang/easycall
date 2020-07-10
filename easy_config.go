package easycall

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/starjiang/elog"
)

type EasyConfig struct {
	rwmutex       *sync.RWMutex
	configPath    string
	configLoaded  bool
	configVersion int
	configName string
	endpoints []string
	container     map[string]string
}

func NewLocalEasyConfig(configPath string) *EasyConfig {
	config := &EasyConfig{}
	config.container = make(map[string]string, 0)
	config.rwmutex = &sync.RWMutex{}
	config.configPath = configPath
	if config.configPath != "" {
		err := config.Load(config.configPath)
		if err != nil {
			elog.Error(err)
		}
	}
	return config
}

func NewRemoteEasyConfig(endpoints []string,configName string) *EasyConfig {
	config := &EasyConfig{}
	config.container = make(map[string]string, 0)
	config.rwmutex = &sync.RWMutex{}
	config.endpoints = endpoints
	config.configName = configName

	err := config.LoadAllConfig()
	if err != nil {
		elog.Error(err)
	}

	return config
}

func (ec *EasyConfig) LoadAllConfig() error {

	err := ec.LoadRemote()
	if err != nil {
		return err
	}

	if !ec.configLoaded {
		go ec.checkRemoteVersion()
		ec.configLoaded = true
	}

	err = ec.LoadLocal()
	if err != nil {
		return err
	}
	return nil
}

func (ec *EasyConfig) checkRemoteVersion() {

	for _ = range time.NewTicker(time.Second * time.Duration(EASYCALL_CONFIG_CHECK_INTERVAL)).C {
		ccVersionPath := EASYCALL_ETCD_CONFIG_PATH + "/" + ec.configName+ "/version"
		remoteVersion, err := ec.etcdGet(ccVersionPath)
		if err != nil {
			elog.Error("check remote config err:", err)
		}
		remoteVersionNum := 0
		remoteVersionNum, err = strconv.Atoi(string(remoteVersion))
		if ec.configVersion < remoteVersionNum {
			elog.Info("reload config...")
			ec.LoadAllConfig()
		}
	}
}

func (ec *EasyConfig) LoadLocal() error {

	localConfigSavePath := EASYCALL_CONFIG_PATH + "/" + ec.configName
	localConfigPath := localConfigSavePath + "/local"

	if !FileIsExist(localConfigSavePath) {
		err := os.MkdirAll(localConfigSavePath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	if FileIsExist(localConfigPath) {
		err := ec.Load(localConfigPath)
		if err != nil {
			elog.Error(err)
		}
	}
	return nil
}

func (ec *EasyConfig) etcdGet(path string) ([]byte,error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   ec.endpoints,
		DialTimeout: ETCD_CONNECT_TIMEOUT * time.Second,
	})
	if err != nil {
		return nil,err
	}
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, path)
	cancel()
	if err != nil {
		return nil,err
	}
	for _, ev := range resp.Kvs {
		return ev.Value,nil
	}
	return nil,errors.New("path not found")
}

func (ec *EasyConfig) LoadRemote() error {

	ccConfigPath := EASYCALL_ETCD_CONFIG_PATH+ "/" + ec.configName+"/data"
	ccVersionPath :=  EASYCALL_ETCD_CONFIG_PATH+ "/" + ec.configName+"/version"

	remoteConfigSavePath := EASYCALL_CONFIG_PATH + "/" + ec.configName

	if !FileIsExist(remoteConfigSavePath) {
		err := os.MkdirAll(remoteConfigSavePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	remoteVersionPath := remoteConfigSavePath + "/version"
	remoteConfigPath := remoteConfigSavePath + "/remote"

	localVersionNum := 0
	remoteVersionNum := 0
	localVerion, err := ioutil.ReadFile(remoteVersionPath)
	if err == nil {
		localVersionNum, err = strconv.Atoi(string(localVerion))
	} else {
		elog.Error(err)
	}

	remoteVersion, err := ec.etcdGet(ccVersionPath)
	if err == nil {
		remoteVersionNum, err = strconv.Atoi(string(remoteVersion))
		ec.configVersion = remoteVersionNum
		err := ioutil.WriteFile(remoteVersionPath, remoteVersion, os.ModePerm)
		if err != nil {
			elog.Error(err)
		}
	} else {
		elog.Error(err)
	}

	if localVersionNum < remoteVersionNum {
		remoteConfigData, err := ec.etcdGet(ccConfigPath)
		if err == nil {
			err := ioutil.WriteFile(remoteConfigPath, remoteConfigData, os.ModePerm)
			if err != nil {
				elog.Error(err)
			}
		} else {
			elog.Error(err)
		}
	}
	if FileIsExist(remoteConfigPath) {
		err := ec.Load(remoteConfigPath)
		if err != nil {
			elog.Error(err)
		}
	}
	return nil
}

func (ec *EasyConfig) Load(fileName string) error {

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		line = strings.TrimSpace(line)
		k, v, flag := ec.getKeyValue(line)
		if flag {
			ec.rwmutex.Lock()
			ec.container[k] = v
			ec.rwmutex.Unlock()
		}
	}
	return nil
}

func (ec *EasyConfig) getKeyValue(line string) (string, string, bool) {
	index := strings.Index(line, "#")
	if index > -1 {
		line = line[:index]
	}

	index = strings.Index(line, "=")
	if index < 1 {
		return "", "", false
	}

	kv := strings.Split(line, "=")

	if len(kv) < 2 {
		return "", "", false
	}

	return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true

}

func (ec *EasyConfig) HasConfig(name string) bool {
	ec.rwmutex.RLock()
	_, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	return ok
}
func (ec *EasyConfig) GetUint64(name string, defaultValue uint64) uint64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		n, _ := strconv.ParseUint(v, 10, 0)
		return n
	} else {
		return defaultValue
	}
}

func (ec *EasyConfig) GetInt64(name string, defaultValue int64) int64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		n, _ := strconv.ParseInt(v, 10, 0)
		return n
	} else {
		return defaultValue
	}
}

func (ec *EasyConfig) GetBool(name string, defaultValue bool) bool {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		n, _ := strconv.ParseBool(v)
		return n
	} else {
		return defaultValue
	}
}

func (ec *EasyConfig) GetFloat64(name string, defaultValue float64) float64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		n, _ := strconv.ParseFloat(v, 0)
		return n
	} else {
		return defaultValue
	}
}

func (ec *EasyConfig) GetString(name string, defaultValue string) string {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return v
	} else {
		return defaultValue
	}
}

func (ec *EasyConfig) GetConfig(prefix string, config interface{}) {
	ec.rwmutex.RLock()
	defer ec.rwmutex.RUnlock()

	value := reflect.ValueOf(config)
	value = value.Elem()
	start := len(prefix)
	for k, v := range ec.container {
		if strings.HasPrefix(k, prefix) {
			index := strings.Index(k, prefix)
			if index > -1 {
				k := k[start:]
				k = strings.ToUpper(k[0:1]) + k[1:]

				field := value.FieldByName(k)
				if field.IsValid() && field.CanSet() {
					if field.Kind() == reflect.Uint || field.Kind() == reflect.Uint8 ||
						field.Kind() == reflect.Uint16 || field.Kind() == reflect.Uint32 ||
						field.Kind() == reflect.Uint64 {
						n, _ := strconv.ParseUint(v, 10, 0)
						field.SetUint(n)
					} else if field.Kind() == reflect.Int || field.Kind() == reflect.Int8 ||
						field.Kind() == reflect.Int16 || field.Kind() == reflect.Int32 ||
						field.Kind() == reflect.Int64 {
						n, _ := strconv.ParseInt(v, 10, 0)
						field.SetInt(n)
					} else if field.Kind() == reflect.Bool {
						n, _ := strconv.ParseBool(v)
						field.SetBool(n)
					} else if field.Kind() == reflect.String {
						field.SetString(v)
					} else if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
						n, _ := strconv.ParseFloat(v, 0)
						field.SetFloat(n)
					} else if field.Kind() == reflect.Slice {

						if _, ok := field.Interface().([]int64); ok {
							list := getInt64List(v)
							field.Set(reflect.ValueOf(list))
						} else if _, ok := field.Interface().([]uint64); ok {
							list := getUint64List(v)
							field.Set(reflect.ValueOf(list))
						} else if _, ok := field.Interface().([]string); ok {
							list := getStringList(v)
							field.Set(reflect.ValueOf(list))
						} else if _, ok := field.Interface().([]float64); ok {
							list := getFloat64List(v)
							field.Set(reflect.ValueOf(list))
						}
					} else if field.Kind() == reflect.Map {
						if _, ok := field.Interface().(map[string]int64); ok {
							kv := getStringInt64Map(v)
							field.Set(reflect.ValueOf(kv))
						} else if _, ok := field.Interface().(map[string]uint64); ok {
							kv := getStringUint64Map(v)
							field.Set(reflect.ValueOf(kv))
						} else if _, ok := field.Interface().(map[string]float64); ok {
							kv := getStringFloat64Map(v)
							field.Set(reflect.ValueOf(kv))
						} else if _, ok := field.Interface().(map[string]bool); ok {
							kv := getStringBoolMap(v)
							field.Set(reflect.ValueOf(kv))
						} else if _, ok := field.Interface().(map[string]string); ok {
							kv := getStringStringMap(v)
							field.Set(reflect.ValueOf(kv))
						}
					}
				}
			}
		}
	}
}

func getStringUint64Map(v string) map[string]uint64 {

	mapValues := make(map[string]uint64, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		kv := strings.Split(value, ":")
		if len(kv) == 2 {
			n, _ := strconv.ParseUint(kv[1], 10, 0)
			mapValues[strings.TrimSpace(kv[0])] = n
		}
	}
	return mapValues
}

func (ec *EasyConfig) GetStringUint64Map(name string) map[string]uint64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getStringUint64Map(v)
	} else {
		return make(map[string]uint64, 0)
	}
}

func getStringStringMap(v string) map[string]string {

	mapValues := make(map[string]string, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		kv := strings.Split(value, ":")
		if len(kv) == 2 {
			mapValues[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return mapValues
}

func (ec *EasyConfig) GetStringStringMap(name string) map[string]string {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getStringStringMap(v)
	} else {
		return make(map[string]string, 0)
	}
}

func getStringInt64Map(v string) map[string]int64 {

	mapValues := make(map[string]int64, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		kv := strings.Split(value, ":")
		if len(kv) == 2 {
			n, _ := strconv.ParseInt(kv[1], 10, 0)
			mapValues[strings.TrimSpace(kv[0])] = n
		}
	}
	return mapValues

}

func (ec *EasyConfig) GetStringInt64Map(name string) map[string]int64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getStringInt64Map(v)
	} else {
		return make(map[string]int64, 0)
	}
}

func getStringFloat64Map(v string) map[string]float64 {

	mapValues := make(map[string]float64, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		kv := strings.Split(value, ":")
		if len(kv) == 2 {
			n, _ := strconv.ParseFloat(kv[1], 0)
			mapValues[strings.TrimSpace(kv[0])] = n
		}
	}
	return mapValues
}

func (ec *EasyConfig) GetStringFloat64Map(name string) map[string]float64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getStringFloat64Map(v)
	} else {
		return make(map[string]float64, 0)
	}
}

func getStringBoolMap(v string) map[string]bool {

	mapValues := make(map[string]bool, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		kv := strings.Split(value, ":")
		if len(kv) == 2 {
			n, _ := strconv.ParseBool(kv[1])
			mapValues[strings.TrimSpace(kv[0])] = n
		}
	}
	return mapValues
}

func (ec *EasyConfig) GetStringBoolMap(name string) map[string]bool {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getStringBoolMap(v)
	} else {
		return make(map[string]bool, 0)
	}
}

func getUint64List(v string) []uint64 {

	intValues := make([]uint64, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		n, _ := strconv.ParseUint(value, 10, 0)
		intValues = append(intValues, n)
	}
	return intValues
}

func (ec *EasyConfig) GetUint64List(name string) []uint64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getUint64List(v)
	} else {
		return make([]uint64, 0)
	}
}

func getFloat64List(v string) []float64 {

	floatValues := make([]float64, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		n, _ := strconv.ParseFloat(value, 0)
		floatValues = append(floatValues, n)
	}
	return floatValues

}

func (ec *EasyConfig) GetFloat64List(name string) []float64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getFloat64List(v)
	} else {
		return make([]float64, 0)
	}
}

func getInt64List(v string) []int64 {

	intValues := make([]int64, 0)
	values := strings.Split(v, ",")
	for _, value := range values {
		n, _ := strconv.ParseInt(value, 10, 0)
		intValues = append(intValues, n)
	}
	return intValues

}

func (ec *EasyConfig) GetInt64List(name string) []int64 {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getInt64List(v)
	} else {
		return make([]int64, 0)
	}
}

func getStringList(v string) []string {
	return strings.Split(v, ",")
}

func (ec *EasyConfig) GetStringList(name string) []string {
	ec.rwmutex.RLock()
	v, ok := ec.container[name]
	ec.rwmutex.RUnlock()
	if ok {
		return getStringList(v)
	} else {
		return make([]string, 0)
	}
}