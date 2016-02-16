package route

import (
	"net/http"
	"os"
	"strings"
)

func download(w http.ResponseWriter, r *http.Request) {
	var filename string = r.FormValue("path")
	if filename == "" {
		http.Error(w, "下载文件不存在", 554)
		return
	}
	if strings.Index(filename, "/") == 0 {
		list := strings.Split(filename, "/")
		filename = strings.Join(list[1:], "/")
	}
	info, err := os.Lstat(filename)
	if err != nil {
		http.Error(w, "下载文件不存在", 554)
		return
	}
	if !info.IsDir() {
		http.ServeFile(w, r, filename)
		return
	}
	http.Error(w, "下载文件不存在", 554)
}
