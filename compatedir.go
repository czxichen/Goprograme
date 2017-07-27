package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const Separator = string(filepath.Separator)

var src, dst, tmp string

func init() {
	flag.StringVar(&src, "s", "", "-s ./ 指定原始目录,不能为空")
	flag.StringVar(&dst, "d", "", "-d 指定目标目录,不能为空")
	flag.StringVar(&tmp, "t", "tmp", "-t tmp 新文件的存放路径")
	flag.Parse()
	if dst == "" || src == "" {
		flag.Usage()
		os.Exit(1)
	}
	src = formatSeparator(src)
	dst = formatSeparator(dst)
	tmp = formatSeparator(tmp)
	err := os.Mkdir(tmp, 0666)
	if err != nil {
		if !os.IsExist(err) {
			fmt.Printf("创建临时目录失败:%s\n", err)
			os.Exit(1)
		}
	}
	src, err = filepath.Abs(src)
	if err != nil {
		fmt.Println("目录转换失败:", err)
	}
}

func main() {
	err := filepath.Walk(src, walk)
	if err != nil {
		fmt.Println(err)
	}
}

func walk(root string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	p := strings.TrimPrefix(root, src)
	_, err = os.Lstat(dst + p)
	if info.IsDir() {
		if err != nil {
			err = copydir(root, tmp, src)
			if err == nil {
				err = filepath.SkipDir
			}
			return err
		}
		return nil

	}
	if err != nil {
		return copyfile(root, tmp+p)
	}
	if !compare(root, dst+p) {
		return copyfile(root, tmp+p)
	}
	return nil
}

func copydir(srcdir, dstdir, sep string) error {
	return filepath.Walk(srcdir, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path := strings.TrimPrefix(root, sep)
		if info.IsDir() {
			return os.Mkdir(dstdir+path, 0666)
		}
		return copyfile(root, dstdir+path)
	})
}

func copyfile(srcfile, dstfile string) error {
	sFile, err := os.Open(srcfile)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(dstfile), 0666)
	dFile, err := os.Create(dstfile)
	if err != nil {
		sFile.Close()
		return err
	}
	_, err = io.Copy(dFile, sFile)
	sFile.Close()
	dFile.Close()
	info, err := os.Lstat(srcfile)
	if err != nil {
		return err
	}
	if err == nil {
		os.Chtimes(dstfile, info.ModTime(), info.ModTime())
	}
	return err
}

func compare(spath, dpath string) bool {
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
	dFile, err := os.Open(dpath)
	if err != nil {
		return false
	}
	b := comparebyte(sFile, dFile)
	sFile.Close()
	dFile.Close()
	return b
}

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

func formatSeparator(path string) string {
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
