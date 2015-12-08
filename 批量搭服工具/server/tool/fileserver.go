package tool

import (
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Template struct {
	Name       string
	Md5        string
	ConfigList []string
}

func Server(ip string) {
	http.Handle("/template/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/", router)
	err := http.ListenAndServe(ip, nil)
	if err != nil {
		fmt.Printf("Start file server error:\n%s\n", err)
		return
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/config":
		Config(w, r)
	default:
		Index(w, r)
	}
}

var TemplateInfo Template

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", GetNow(), r.RemoteAddr)
	gob.NewEncoder(w).Encode(TemplateInfo)
}
func Config(w http.ResponseWriter, r *http.Request) {
	str := r.RemoteAddr
	remoteIP := strings.Split(str, ":")[0]
	info := Matching(remoteIP)
	valueList := Split(info)
	Key_Map := Merge(valueList)
	key := r.URL.Query().Get("key")
	w.Header().Set("path", Temp_Path[key])
	ExecuteReplace(w, key, Key_Map)
}

func GetNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Md5(path string) string {
	File, err := os.Open(path)
	if err != nil {
		fmt.Printf("Check md5 error:\n%s\n", err)
		return ""
	}
	m := md5.New()
	io.Copy(m, File)
	return fmt.Sprintf("%X", string(m.Sum([]byte{})))
}
