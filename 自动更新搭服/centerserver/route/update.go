package route

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
)

func index(w http.ResponseWriter) {
	io.WriteString(w, "Welcome To AutoServer!")
}

func update(w http.ResponseWriter, r *http.Request) {
	//如果根已经对身份进行验证,次出就不必验证.
	//	if !auth(r) {
	//		io.WriteString(w, "无法识别身份")
	//	}
	var (
		path      string
		gameid    string = r.FormValue("gameid")
		updateDir string = r.FormValue("dir")
	)
	if gameid == "" {
		http.Error(w, "无法获取到游戏id", 550)
		Log.PrintfE("IP->%s 无法获取到游戏id: %d\n", r.RemoteAddr, 550)
		return
	}
	//如果请求指定更新的目录,则使用请求的路径.
	if updateDir == "" {
		Log.PrintfI("%s %s\n", r.RemoteAddr, DownLoadPath+gameid+"/")
		updateDir = getNewDir(DownLoadPath + gameid + "/")
		if len(updateDir) == 0 {
			http.Error(w, "未找到可更新目录", 551)
			Log.PrintfE("IP->%s 未找到可更新目录: %s\n", r.RemoteAddr, updateDir)
			return
		}
	}
	path = DownLoadPath + gameid + "/" + updateDir + "/"
	//这里是更新请求的路径需要更改
	l := parseUpdatedir(path)
	if len(l) == 0 {
		http.Error(w, "未找到可更新的文件", 552)
		Log.PrintfE("IP->%s 未找到可更新的文件: %s\n", r.RemoteAddr, path)
		return
	}
	sort.Sort(l)
	var result updateResult = updateResult{path, ([]string)(l)}
	buf, err := json.Marshal(result)
	if err != nil {
		//定义Json编码失败,状态码552.
		http.Error(w, "Json编码失败", 553)
		Log.PrintfE("IP->%s %s\n", r.RemoteAddr, err)
		return
	}
	w.Write(buf)
}
