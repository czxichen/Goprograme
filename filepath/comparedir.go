package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var zipdir bool
var src, dsc string
var diffpath = "DifferenceFile/"

func main() {
	flag.StringVar(&src, "s", "", "-s 原始目录,不能为空")
	flag.StringVar(&dsc, "d", "", "-d 目标目录,不能为空")
	flag.StringVar(&diffpath, "D", "DifferenceFile/", "-D 不一致文件存放目录")
	flag.BoolVar(&zipdir, "z", false, "-z 或者 -z true,加此参数则会将不一致的文件压缩")
	flag.Parse()
	if src == "" || dsc == "" {
		flag.Usage()
		os.Exit(1)
	}
	src = filepath.ToSlash(src)
	dsc = filepath.ToSlash(dsc)
	diffpath = filepath.ToSlash(diffpath)
	if !strings.HasSuffix(src, "/") {
		src += "/"
	}
	if !strings.HasSuffix(dsc, "/") {
		dsc += "/"
	}
	if !strings.HasSuffix(diffpath, "/") {
		diffpath += "/"
	}

	filepath.Walk(src, walk)
	if zipdir {
		err := Zip(diffpath, time.Now().Format("2006-01-02")+".zip")
		if err != nil {
			fmt.Println("压缩错误:", err.Error())
		}
	}
}

func walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	path = strings.TrimPrefix(filepath.ToSlash(path), src)
	Handler(path)
	return nil
}

func Handler(path string) {
	if !Compare(src+path, dsc+path) {
		err := CopyFile(src+path, diffpath+path)
		if err != nil {
			fmt.Println("拷贝文件失败:", err.Error())
		}
	}
}

func CopyFile(sfile, dfile string) error {
	err := copyFile(sfile, dfile)
	if err != nil {
		os.Remove(dfile)
		return err
	}
	sinfo, _ := os.Lstat(sfile)
	os.Chmod(dfile, sinfo.Mode())
	os.Chtimes(dfile, sinfo.ModTime(), sinfo.ModTime())
	return nil
}

func copyFile(sfile, dfile string) error {
	sFile, err := os.Open(sfile)
	if err != nil {
		return err
	}
	defer sFile.Close()

	dir := filepath.Dir(dfile)
	err = os.MkdirAll(dir, 0666)
	if err != nil {
		return err
	}
	dFile, err := os.Create(dfile)
	if err != nil {
		return err
	}
	defer dFile.Close()

	_, err = io.Copy(dFile, sFile)
	return err
}

func Compare(spath, dpath string) bool {
	sinfo, err := os.Lstat(spath)
	if err != nil {
		return false
	}
	dinfo, err := os.Lstat(dpath)
	if err != nil {
		return false
	}
	if sinfo.Size() != dinfo.Size() || !sinfo.ModTime().Equal(dinfo.ModTime()) {
		return false
	}
	return comparefile(spath, dpath)
}

func comparefile(spath, dpath string) bool {
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
func comparebyte(sfile *os.File, dfile *os.File) bool {
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

const zone int64 = +8

func Zip(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			header.Method = zip.Deflate
		}
		header.SetModTime(time.Unix(info.ModTime().Unix()+(zone*60*60), 0))
		header.Name = path
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}
