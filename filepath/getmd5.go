package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/czxichen/AutoWork/tools/md5"
	"github.com/czxichen/AutoWork/tools/split"
)

var (
	raddr, laddr        string
	sdir, ddir, exclude string
	passName            []string
	result              string = "result/"
)

func init() {
	flag.StringVar(&ddir, "d", "", "-d 指定要匹配的目录")
	flag.StringVar(&sdir, "s", "", "-s 指定要读取的目录")
	flag.StringVar(&raddr, "p", "", "-p 指定原始目录的IP和端口")
	flag.StringVar(&laddr, "l", ":1789", "-l 127.0.0.1:1789 指定监听的端口")
	flag.StringVar(&exclude, "v", "", "-v log,txt 指定排除的后缀文件")
	flag.Parse()
}

func main() {
	if sdir != "" && laddr != "" {
		Server()
		return
	}
	if ddir != "" && raddr != "" {
		Client(raddr)
		return
	}
	flag.Usage()
}

func Server() {
	sdir = filepath.ToSlash(sdir)
	if !strings.HasSuffix(sdir, "/") {
		sdir += "/"
	}
	if exclude != "" {
		passName = strings.Split(exclude, ",")
	}

	os.MkdirAll(result, 0666)
	Walk(sdir)
	http.HandleFunc("/", Router)
	http.ListenAndServe(laddr, nil)
}

func Router(w http.ResponseWriter, r *http.Request) {
	log.Printf("远端地址:%s\t访问的路径:%s\n", r.RemoteAddr, r.URL.Path)
	defer r.Body.Close()
	switch r.URL.Path {
	case "/":
		File, err := os.Open("md5_list.txt")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, File)
		File.Close()
	case "/result":
		file, err := os.Create(result + strings.Split(r.RemoteAddr, ":")[0] + ".txt")
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
		io.Copy(file, r.Body)
		file.Close()
	case "/flushmd5":
		Walk(sdir)
		fmt.Fprintln(w, "flush md5 list ok")
	}
}

func Walk(dir string) {
	File, err := os.Create("md5_list.txt")
	if err != nil {
		log.Println("创建md5列表文件失败:", err.Error())
		return
	}
	defer File.Close()

	err = filepath.Walk(dir, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if Exclude(root) {
			return nil
		}
		m5, err := md5.Md5(root)
		if err != nil {
			log.Printf("读取文件:%s的md5失败,错误信息:\n", root, err.Error())
			return err
		}
		root = strings.TrimPrefix(filepath.ToSlash(root), dir)
		fmt.Fprintln(File, root, m5)
		return nil
	})
	if err != nil {
		log.Println("遍历文件夹出错:", err.Error())
	}
	log.Println("遍历获取md5完成")
}

func Exclude(Suffix string) bool {
	for _, name := range passName {
		if strings.HasSuffix(Suffix, name) {
			return true
		}
	}
	return false
}

func Client(ip string) {
	ddir = filepath.ToSlash(ddir)
	if !strings.HasSuffix(ddir, "/") {
		ddir += "/"
	}

	resp, err := http.Get("http://" + ip)
	if err != nil {
		fmt.Println("连接远端出错:", err.Error())
		return
	}
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, resp.Body)
	resp.Body.Close()
	File, err := os.Create("cmd5_list.txt")
	if err != nil {
		fmt.Println("创建结果文件失败:", err.Error())
		return
	}
	defer File.Close()
	err = Compare(buf, File, ddir)
	if err != nil {
		fmt.Println(err)
	}
	File.Sync()
	File.Seek(0, 0)
	err = client(File, ip)
	if err != nil {
		fmt.Println("上传结果出错:", err.Error())
	}
	os.Remove("cmd5_list.txt")
}

func Compare(r io.Reader, w io.Writer, dst string) error {
	rd := bufio.NewReader(r)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		path_md5 := split.Split(string(line))
		if len(path_md5) != 2 {
			continue
		}
		m5, err := md5.Md5(dst + path_md5[0])
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(w, path_md5[0], path_md5[1], "'File_not_is_exist'")
			} else {
				fmt.Fprintln(w, path_md5[0], path_md5[1], err.Error())
			}
			continue
		}
		if path_md5[1] != m5 {
			fmt.Fprintln(w, path_md5[0], path_md5[1], m5)
		}
	}
	return nil
}

func client(r io.Reader, ip string) error {
	resp, err := http.Post("http://"+ip+"/result", "application/octet-stream", r)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
