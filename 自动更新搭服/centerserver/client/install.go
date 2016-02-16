package client

import (
	"centerserver/tools"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Install() {
	url := getUrl() + "/config?gameid=" + Config.GameId
	Log.PrintfI("%v\n", url)
	resp, err := client().Get(url)
	if err != nil || resp.StatusCode != 200 {
		Log.PrintfE("request error status code %d %v\n", resp.StatusCode, err)
		return
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.PrintfE("read body error %v\n", err)
		return
	}
	Log.PrintfI("%v\n", string(buf))
	var configInfo responeconfig
	if err = json.Unmarshal(buf, &configInfo); err != nil {
		Log.PrintfE("Unmarshal json error %v\n", err)
		return
	}
	Log.PrintfI("%v\n", buf)
	Log.PrintfI("%v\n", configInfo)
	getServerTemplate(configInfo.TempPath)
	for k, v := range configInfo.PathConfig {
		createConfigFile(k, Config.Primary, v)
	}
}

func getServerTemplate(path string) {
	url := getUrl() + "/download?path=" + path
	Log.PrintfI("start download server template %v\n", url)
	resp, err := client().Get(url)
	if err != nil {
		Log.PrintfE("%v\n", err)
		return
	}
	defer resp.Body.Close()
	name := filepath.Base(path)
	//应该判断下Config.TmpDir 是否包含"/"
	File, err := os.Create(Config.TmpDir + "/" + name)
	if err != nil {
		Log.PrintfF("%v\n", err)
		return
	}
	io.Copy(File, resp.Body)
	File.Close()
	if Config.CheckMd5 {
		if !checkMd5(Config.TmpDir + "/" + name) {
			Log.PrintfF("%v\n", "md5 different")
		}
	}
	tools.Unzip(Config.TmpDir+"/"+name, Config.ProgramHome, Log)
}

func createConfigFile(key, serverid string, path []string) {
	url := getUrl() + "/install?key=" + key + "&gameid=" + Config.GameId
	if serverid != "" {
		url = url + "&serverid=" + serverid
	}
	resp, err := client().Get(url)
	if err != nil {
		Log.PrintfE("%v\n", err)
		return
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.PrintfE("%v\n", err)
		return
	}
	for _, v := range path {
		Log.PrintfI("start create %s config.", v)
		File, err := os.Create(Config.ProgramHome + "/" + v)
		if err != nil {
			Log.PrintfW("%v\n", err)
			continue
		}
		defer File.Close()
		File.Write(buf)
	}
}

func checkMd5(path string) bool {
	Log.PrintfI("start check %v md5.\n", path)
	if !tools.CheckValidZip(path) {
		Log.PrintfE("server template %v isn't a valid zip file.", path)
		return false
	}
	name := strings.Split(filepath.Base(path), ".")[0]
	m := strings.Split(name, "_")
	if len(m) != 2 && len(m) != 3 {
		Log.PrintfE("parse server teamplate name faild.%v\n", m)
		return false
	}
	str, err := tools.Md5(path)
	if err != nil {
		Log.PrintfE("get md5 error %v\n", err)
		return false
	}
	Log.PrintfI("%v %v\n", path, str)
	return strings.ToUpper(m[1]) == str
}
