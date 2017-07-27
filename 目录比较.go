package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type compareConfig struct {
	dpath   string
	spath   string
	quick   bool
	diff    string
	output  string
	split   string
	getmd5  bool
	md5list string
	laddr   string
}

var compareCFG compareConfig

func init() {
	flag.StringVar(&compareCFG.spath, "s", "", `-s="server" 指定源目录或文件`)
	flag.StringVar(&compareCFG.dpath, "d", "", `-d="server_back" 指定目标目录或文件`)
	flag.StringVar(&compareCFG.diff, "c", "", `-c="diff" 把源目录不匹配的文件提出来,为空则不拷贝`)
	flag.StringVar(&compareCFG.output, "o", "", `指定匹配的结果输出文件,不指定则输出到标准输出`)
	flag.StringVar(&compareCFG.split, "S", "	", `-S=" " 验证指定路径文件的md5值的时候指定md5和路径的分隔符`)
	flag.StringVar(&compareCFG.md5list, "F", "", `-F="md5list" 验证指定路径文件的md5值,每行一条数据,md5码和文件路径,结合-d使用`)
	flag.BoolVar(&compareCFG.quick, "q", true, `-q=false 快速模式即用文件内容判断文件是否一致,如果为false则使用md5比较`)
	flag.BoolVar(&compareCFG.getmd5, "m", false, "-m true 获取指定目录下所有文件的md5,结合-s指定目录使用")
	flag.StringVar(&compareCFG.laddr, "l", "", `-l="192.168.1.2:80" 通过http协议共享md5列表,结合-m=true使用`)
	flag.Parse()
}

func main() {
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		return
	}
	var w = os.Stdout
	if compareCFG.output != "" {
		var err error
		w, err = os.Create(compareCFG.output)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if compareCFG.getmd5 {
		if compareCFG.spath != "" {
			if compareCFG.laddr != "" {
				err := server(compareCFG.laddr, compareCFG.spath)
				if err != nil {
					log.Println(err)
				}
				return
			}
			err := GetMd5(compareCFG.spath, w)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("必须指定目录路径")
		}
		return
	}

	if compareCFG.spath != "" && compareCFG.dpath != "" {
		comparePath(compareCFG.spath, compareCFG.dpath, compareCFG.diff, w, compareCFG.quick)
		return
	}

	if compareCFG.md5list != "" && compareCFG.dpath != "" {
		var r io.ReadCloser
		var err error
		defer func() {
			if r != nil {
				r.Close()
			}
		}()

		if strings.HasPrefix(compareCFG.md5list, "http") {
			//	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			resp, err := http.Get(compareCFG.md5list)
			if err != nil {
				log.Println(err)
				return
			}
			r = resp.Body
		} else {
			r, err = os.Open(compareCFG.md5list)
			if err != nil {
				log.Println(err)
				return
			}
		}
		compareFromFile(r, compareCFG.dpath, compareCFG.split, w)
		return
	}
	flag.PrintDefaults()
}

func compareFromFile(File io.Reader, dir, split string, w io.Writer) {
	dir = formatSeparator(dir)
	buf := bufio.NewReader(File)
	h := md5.New()
	var m, dfile, md string
	var count int = 0
	for {
		count += 1
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		list := bytes.Split(line, []byte(split))
		if len(list) != 2 {
			log.Printf("第%d行数据无效,确认分隔符\n", count)
			continue
		}
		dfile = string(bytes.TrimSpace(list[1]))
		md, err = getmd5(h, dir+dfile)
		if err != nil {
			fmt.Fprintf(w, "文件不存在:\t%s\n", dfile)
			continue
		}

		m = string(bytes.TrimSpace(list[0]))
		if md != m {
			fmt.Fprintf(w, "内容不一致:\t%s\n", dfile)
		}
	}
}

func comparePath(spath, dpath, diff string, w io.Writer, quick bool) {
	sinfo, err := os.Lstat(spath)
	if err != nil {
		log.Println(err)
		return
	}
	dinfo, err := os.Lstat(dpath)
	if err != nil {
		log.Println(err)
		return
	}

	if !sinfo.IsDir() && !dinfo.IsDir() {
		if comparefile(spath, dpath, quick) {
			fmt.Fprintf(w, "内容一致:\t%s\n", dinfo.Name())
		} else {
			fmt.Fprintf(w, "内容不一致:\t%s\n", dinfo.Name())
		}
		return
	}

	if !(sinfo.IsDir() || dinfo.IsDir()) {
		log.Println("原路径和目标路径必须同时为文件或者同时为目录")
		return
	}

	if diff != "" {
		diff = formatSeparator(diff)
		os.RemoveAll(diff)
		os.MkdirAll(diff, 0666)
	}

	spath = formatSeparator(spath)
	dpath = formatSeparator(dpath)

	filepath.Walk(spath, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		path := strings.TrimPrefix(root, spath)
		dfile := dpath + path
		dinfo, err := os.Lstat(dfile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(w, "文件不存在:\t%s\n", path)
				if diff != "" {
					err = copyFile(root, diff+path)
					if err != nil {
						log.Printf("拷贝文件错误:%s\n", err)
					} else {
						log.Printf("拷贝文件成功:%s\n", path)
					}
				}
			}
			return nil
		}

		if info.Size() == dinfo.Size() {
			if comparefile(root, dfile, quick) {
				fmt.Fprintf(w, "内容一致:\t%s\n", path)
				return nil
			}
		}
		fmt.Fprintf(w, "内容不一致:\t%s\n", path)
		if diff != "" {
			err = copyFile(root, diff+path)
			if err != nil {
				log.Printf("拷贝文件错误:%s\n", err)
			} else {
				log.Printf("拷贝文件成功:%s\n", path)
			}
		}
		return nil
	})
	return
}

func comparefile(spath, dpath string, quick bool) bool {
	if !quick {
		h := md5.New()
		smd5, err := getmd5(h, spath)
		if err != nil {
			return false
		}
		dmd5, err := getmd5(h, dpath)
		if err != nil {
			return false
		}
		return smd5 == dmd5
	}

	sFile, err := os.Open(spath)
	if err != nil {
		return false
	}
	defer sFile.Close()
	dFile, err := os.Open(dpath)
	if err != nil {
		return false
	}
	defer dFile.Close()
	return comparebyte(sFile, dFile)
}

//下面可以代替md5比较.
func comparebyte(sfile io.Reader, dfile io.Reader) bool {
	var sbyte []byte = make([]byte, 512)
	var dbyte []byte = make([]byte, 512)
	var serr, derr error
	for {
		_, serr = sfile.Read(sbyte)
		_, derr = dfile.Read(dbyte)
		if serr != nil || derr != nil {
			if serr != derr {
				return false
			}
			if serr == io.EOF {
				break
			}
		}
		if bytes.Equal(sbyte, dbyte) {
			continue
		}
		return false
	}
	return true
}

func copyFile(spath, dpath string) error {
	err := copyfile(spath, dpath)
	if err != nil {
		return err
	}
	info, err := os.Lstat(spath)
	if err != nil {
		return err
	}
	os.Chmod(dpath, info.Mode())
	os.Chtimes(dpath, info.ModTime(), info.ModTime())
	return nil
}

func copyfile(spath, dpath string) error {
	basedir := filepath.Dir(dpath)
	err := os.MkdirAll(basedir, 0666)
	if err != nil {
		return err
	}
	dFile, err := os.Create(dpath)
	if err != nil {
		return err
	}
	defer dFile.Close()
	sFile, err := os.Open(spath)
	if err != nil {
		return err
	}
	defer sFile.Close()
	_, err = io.Copy(dFile, sFile)
	return err
}

func formatSeparator(path string) string {
	Separator := string(filepath.Separator)
	if Separator != "/" {
		path = strings.Replace(path, "/", Separator, -1)
	} else {
		path = strings.Replace(path, "\\", Separator, -1)
	}
	if !strings.HasSuffix(path, Separator) {
		path += Separator
	}
	return path
}

func GetMd5(path string, w io.Writer) error {
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("路径必须是目录")
	}
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		path += string(filepath.Separator)
	}
	h := md5.New()
	return filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dir := strings.TrimPrefix(root, path)
		m, err := getmd5(h, root)
		if err != nil {
			log.Println("计算md5失败:", err.Error())
			return nil
		}
		if filepath.Separator == '\\' {
			dir = strings.Replace(dir, "\\", "/", -1)
		}

		fmt.Fprintf(w, "%s\t%s\n", m, dir)
		return nil
	})
}

func getmd5(md5hash hash.Hash, path string) (string, error) {
	File, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer File.Close()
	md5hash.Reset()
	_, err = io.Copy(md5hash, File)
	if err != nil {
		return "", err
	}
	result := make([]byte, 0, 32)
	result = md5hash.Sum(result)
	return hex.EncodeToString(result), nil
}

/*
func server(laddr string, dirpath string) error {
	var list []byte
	flush := func() error {
		buf := bytes.NewBuffer(nil)
		if err := GetMd5(dirpath, buf); err != nil {
			return err
		}
		list = buf.Bytes()
		return nil
	}

	if err := flush(); err != nil {
		return err
	}

	route := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("远程地址:%s 访问路径:%s\n", r.RemoteAddr, r.URL.Path)
		defer r.Body.Close()
		switch r.URL.Path {
		default:
			fmt.Fprintln(w, `<a href="/getmd5">查看MD5列表</a><br><a href="/flush">刷新MD5列表</a>`)
		case "/getmd5":
			fmt.Fprintf(w, "%s", list)
		case "/flush":
			if err := flush(); err != nil {
				fmt.Fprintf(w, "刷新md5列表失败,%s\n", err)
			} else {
				fmt.Fprintln(w, "刷新md5列表成功")
			}
		}
	}

	http.HandleFunc("/", route)
	return http.ListenAndServe(laddr, nil)
}
*/
