package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

type Config struct {
	IP             string `json:ip`
	Mode           string `json:mode`
	CertFile       string `json:certfile`
	KeyFile        string `json:keyfile`
	LogPath        string `json:logpath`
	HttpLogPath    string `json:logpath`
	FileServerPath string `json:fileserverpath`
}

var (
	cfg  *Config
	lock *sync.RWMutex = new(sync.RWMutex)
)

func ParseConfig(configpath string) {
	buf, err := ioutil.ReadFile(configpath)
	if err != nil {
		if cfg == nil {
			if Log != nil {
				Log.PrintfF("读取配置文件失败 %s\n", err)
				return
			}
			fmt.Printf("读取配置文件失败 %s\n", err)
		}
	}
	lock.Lock()
	defer lock.Unlock()
	cfg = new(Config)
	err = json.Unmarshal(buf, cfg)
	if err != nil {
		if cfg.IP == "" {
			if Log != nil {
				Log.PrintfF("解析配置文件失败 %s\n", err)
			} else {
				fmt.Printf("解析配置文件失败 %s\n", err)
			}
		}
		if Log != nil {
			Log.PrintfW("解析配置文件失败,配置未更改. %s\n", "")
			return
		}
		fmt.Printf("解析配置文件失败,配置未更改. %s\n", "")
	}
}

func GetConfig() *Config {
	lock.RLock()
	defer lock.RUnlock()
	return cfg
}
