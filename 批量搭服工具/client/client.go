package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

type Template struct {
	Name       string
	Md5        string
	ConfigList []string
}

var url string
var homedir string
var Temp_Path map[string]string = make(map[string]string)
var HeadList []string
var Replational []string

func main() {
	flag.StringVar(&url, "u", "http://127.0.0.1:1789", "Specify server address.")
	flag.StringVar(&homedir, "d", "/data/gamehome", "Specify home directory.")
	L := flag.Bool("l", false, "local generation config.")
	cfgTempRelational := flag.String("c", "CfgTempRelational.ini", "Specify CfgTempRelational config")
	relationalTable := flag.String("r", "RelationalTable.ini", "Specify RelationalTable config.")
	name := flag.String("n", "server.zip", "Specify server package name.")
	flag.Parse()
	if !strings.HasSuffix(homedir, "/") {
		homedir = homedir + "/"
	}
	if *L {
		if Unzip(*name, homedir) {
			LocalReplace(*cfgTempRelational, *relationalTable)
			return
		}
		return
	}
	fmt.Printf("%s Start request template info.From %s\n", GetNow(), url)
	req, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s Request info error\n", GetNow())
		return
	}
	var TemplateInfo Template
	func(t *Template) {
		defer req.Body.Close()
		err := gob.NewDecoder(req.Body).Decode(t)
		if err != nil {
			fmt.Printf("%s Parse template info:\n%s\n", GetNow(), err)
			os.Exit(1)
		}
	}(&TemplateInfo)
	fmt.Printf("%s Parse Template info is OK.\n", GetNow())
	fmt.Printf("%s Print Template Info:\n%s\n", GetNow(), TemplateInfo)
	if Download(TemplateInfo) {
		if !Unzip(TemplateInfo.Name, homedir) {
			fmt.Printf("%s Unzip file error.\n", GetNow())
			os.Exit(2)
		}
	}
	ConfigDownload(TemplateInfo.ConfigList)
}

func LocalReplace(CfgTempRelational, RelationalTable string) {
	fmt.Printf("start parse %s\n", CfgTempRelational)
	GetPathConfig(CfgTempRelational)
	fmt.Printf("start parse %s\n", RelationalTable)
	ParseServerConfig(RelationalTable)
	localip := GetLocalIP()
	fmt.Printf("get local ip %s\n", localip)
	for _, v := range localip {
		info := Matching(v)
		if len(info) > 0 {
			valueList := Split(info)
			fmt.Printf("Match values %s\n", valueList)
			Key_Map := Merge(valueList)
			for k, v := range Temp_Path {
				fmt.Printf("create config %s%s\n", homedir, v)
				File, err := os.Create(fmt.Sprintf("%s%s", homedir, v))
				if err != nil {
					fmt.Printf("%s create %s faild\n%s\n", GetNow(), v, err)
					os.Exit(1)
				}
				ExecuteReplace(File, k, Key_Map)
			}
			return
		}
	}
}

func ConfigDownload(TemplateList []string) {
	for _, v := range TemplateList {
		resp, err := http.Get(fmt.Sprintf("%s/config?key=%s", url, v))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		path := homedir + resp.Header.Get("path")
		fmt.Printf("%s Start create %s file\n", GetNow(), path)
		File, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		io.Copy(File, resp.Body)
	}
	fmt.Printf("%s Create config is Ok.\n", GetNow())
}

func Download(TemplateInfo Template) bool {
	_, err := os.Lstat(TemplateInfo.Name)
	if err == nil {
		if Md5(TemplateInfo.Name) == TemplateInfo.Md5 {
			fmt.Printf("%s local is exist %s\n", GetNow(), TemplateInfo.Name)
			return true
		}
	}
	fmt.Printf("%s Start download %s\n", GetNow(), TemplateInfo.Name)
	u := fmt.Sprintf("%s/template/%s", url, TemplateInfo.Name)
	req, err := http.Get(u)
	if err != nil {
		fmt.Printf("%s Open %s error:\n%s\n", GetNow(), u, err)
		return false
	}
	defer req.Body.Close()
	File, err := os.Create(TemplateInfo.Name)
	if err != nil {
		fmt.Printf("%s Create %s error:\n%s\n", GetNow(), TemplateInfo.Name, err)
		return false
	}
	io.Copy(File, req.Body)
	if Md5(TemplateInfo.Name) != TemplateInfo.Md5 {
		fmt.Printf("%s Check md5 faild!\n", GetNow())
		return false
	}
	return true
}

func Md5(path string) string {
	fmt.Printf("%s Check md5 %s\n", GetNow(), path)
	File, err := os.Open(path)
	if err != nil {
		fmt.Printf("%s Check md5 error:\n%s\n", GetNow(), err)
		return ""
	}
	m := md5.New()
	io.Copy(m, File)
	return fmt.Sprintf("%X", string(m.Sum([]byte{})))
}

func Unzip(filename, dir string) bool {
	fmt.Printf("%s Unzip to %s\n", GetNow(), dir)
	File, err := zip.OpenReader(filename)
	if err != nil {
		fmt.Printf("%s Open zip faild:\n%s\n", GetNow(), err)
		return false
	}
	defer File.Close()
	for _, v := range File.File {
		v.Name = fmt.Sprintf("%s%s", dir, v.Name)
		info := v.FileInfo()
		if info.IsDir() {
			err := os.MkdirAll(v.Name, 0644)
			if err != nil {
				fmt.Printf("%s Create direcotry %s faild:\n%s\n", GetNow(), v.Name, err)
				return false
			}
			continue
		}
		srcFile, err := v.Open()
		if err != nil {
			fmt.Printf("%s Read from zip faild:\n%s\n", GetNow(), err)
			return false
		}
		defer srcFile.Close()
		newFile, err := os.Create(v.Name)
		if err != nil {
			fmt.Printf("%s Create file faild:\n%s\n", GetNow(), err)
			return false
		}
		io.Copy(newFile, srcFile)
		newFile.Close()
	}
	return true
}

func ExecuteReplace(w *os.File, temp_path string, funcs map[string]string) error {
	fmt.Println(funcs)
	T := template.New("")
	buf, err := ioutil.ReadFile(temp_path)
	if err != nil {
		return err
	}
	T, err = T.Parse(string(buf))
	if err != nil {
		return err
	}
	err = T.Execute(w, funcs)
	if err != nil {
		return err
	}
	return nil
}

func GetPathConfig(path string) {
	File, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	defer File.Close()
	M := make(map[string]string)
	Buf := bufio.NewReader(File)
	var linenum int = 1
	for {
		line, _, err := Buf.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
			os.Exit(5)
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		list := bytes.Split(line, []byte("="))
		if len(list) != 2 {
			fmt.Printf("check config %s ,line %d\n", path, linenum)
			os.Exit(6)
		}
		key := string(bytes.TrimSpace(list[0]))
		value := string(bytes.TrimSpace(list[1]))
		TestPath(key)
		M[key] = value
		linenum++
	}
	if len(M) < 1 {
		fmt.Printf("config %s can't emptey!\n", path)
		os.Exit(7)
	}
	Temp_Path = M
}
func TestPath(path string) {
	info, err := os.Lstat(path)
	if err != nil {
		fmt.Printf("check %s .error_info :%s", path, err)
		os.Exit(9)
	}
	if info.IsDir() {
		fmt.Printf("check %s is directory", path)
		os.Exit(10)
	}
	return
}
func Merge(list []string) map[string]string {
	ExecuteReplaceValue := make(map[string]string)
	for k, v := range HeadList {
		ExecuteReplaceValue[v] = list[k]
	}
	return ExecuteReplaceValue
}

func ParseServerConfig(path string) {
	File, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer File.Close()
	buf := bufio.NewReader(File)
	var linenum int = 1
	for i := 0; i < 1001; i++ { //just init top 1000
		line, _, err := buf.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
			os.Exit(1)
		}
		if len(HeadList) == 0 {
			list := Split(string(line))
			if len(list) > 0 {
				HeadList = list
			}
			continue
		}
		list := Split(string(line))
		if len(list) != len(HeadList) {
			fmt.Printf("Line %d parse error.", linenum)
			continue
		}
		Replational = append(Replational, string(line))
	}
	if len(Replational) <= 0 {
		fmt.Println("read config error.")
		os.Exit(1)
	}
}

func Split(str string) []string {
	var l []string
	list := strings.Split(str, " ")
	for _, v := range list {
		if len(v) == 0 {
			continue
		}
		if strings.Contains(v, "	") {
			list := strings.Split(v, "	")
			for _, v := range list {
				if len(v) == 0 {
					continue
				}
				l = append(l, v)
			}
			continue
		}
		l = append(l, v)
	}
	return l
}
func GetLocalIP() []string {
	list, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	var l []string
	for _, v := range list {
		if strings.Contains(v.String(), ":") {
			continue
		}
		ip := strings.Split(v.String(), "/")
		if len(ip) != 2 {
			continue
		}
		l = append(l, ip[0])
	}
	return l
}
func Matching(srcip string) string {
	for _, v := range Replational {
		if strings.Contains(v, srcip) {
			return v
		}
	}
	return ""
}

func GetNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
