package client

import (
	"centerserver/tools"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func client() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func getUrl() string {
	if Config.RequestMode != "http" && Config.RequestMode != "https" {
		Log.PrintfE("%s\n", "request mode must http or https")
		os.Exit(-3)
	}
	if !strings.HasSuffix(Config.MasteUrl, "/") {
		return Config.RequestMode + "://" + Config.MasteUrl + "/"
	}
	return Config.RequestMode + "://" + Config.MasteUrl
}

func Update() {
	var m url.Values = make(url.Values)
	m.Set("gameid", Config.GameId)
	if updateDir != "" {
		m.Set("dir", updateDir)
	}
	url := getUrl()
	resq, err := client().PostForm(url+"update", m)
	if err != nil {
		Log.PrintfE("%s\n", err)
		return
	}
	if resq.StatusCode != 200 {
		Log.PrintfE("ErrorCode: %d\n", resq.StatusCode)
		return
	}
	defer resq.Body.Close()
	var x updateResult
	buf, _ := ioutil.ReadAll(resq.Body)
	err = json.Unmarshal(buf, &x)
	if err != nil {
		return
	}
	tmpdir := Config.TmpDir + "/" + filepath.Base(x.Path)
	err = os.MkdirAll(tmpdir, 0644)
	if err != nil {
		Log.PrintfE("%s\n", err)
		return
	}
	var updatelist []string
	for _, v := range x.Files {
		url := fmt.Sprintf("%sdownload?path=%s%s", url, x.Path, v)
		Log.PrintfI("%s\n", url)
		b, err := tools.Wget(url, tmpdir+"/"+v)
		if !b {
			Log.PrintfF("%s\n", err)
		}
		if !tools.CheckValidZip(tmpdir + "/" + v) {
			Log.PrintfW("%v isn't a valid zip file.\n", tmpdir+"/"+v)
			continue
		}
		if !checkMd5(tmpdir + "/" + v) {
			Log.PrintfF("check %v md5 faild.\n", tmpdir+"/"+v)
		}
		updatelist = append(updatelist, tmpdir+"/"+v)
	}
	Log.PrintfI("update files %v\n", updatelist)
	for _, zipfile := range updatelist {
		Log.PrintfI("start unzip %v\n", zipfile)
		if err := tools.Unzip(zipfile, Config.ProgramHome, Log); err != nil {
			Log.PrintfE("%v\n", err)
		}
	}
}
