package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {
	var i FindFiles = FindFiles{"AutoUpdate", false, true}
	//	list, err := i.RegFindFile("log.*")
	//	fmt.Println(list, err)
	list, err := i.DateFindFile(30)
	fmt.Println(list, err)
}

type FindFiles struct {
	Path     string `json:path`
	FullDir  bool   `json:fulldir`
	MatchDir bool   `json:matchdir`
}

func (self FindFiles) DateFindFile(date int64) ([]string, error) {
	date = date * 24 * 60 * 60
	var less bool
	switch {
	case date <= 0:
		date = time.Now().Unix() + date
		less = true
	case date > 0:
		date = time.Now().Unix() - date
		less = false
	}
	return datewalk(date, less, self.MatchDir, self.Path)
}

func (self FindFiles) RegFindFile(reg string) ([]string, error) {
	if strings.Index(reg, "*") == 0 {
		reg = "." + reg
	} else {
		reg = "^" + reg
	}
	reg += "$"
	Reg, err := regexp.Compile(reg)
	if err != nil {
		return []string{}, nil
	}
	if self.FullDir {
		return namewalk(Reg, self.MatchDir, self.Path)
	}
	var list []string
	infos, err := readDir(self.Path)
	if err != nil {
		return list, nil
	}
	path := filepath.ToSlash(self.Path)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	for _, v := range infos {
		if Reg.MatchString(v.Name()) {
			list = append(list, path+v.Name())
		}
	}
	return list, nil
}
func datewalk(date int64, less bool, matchdir bool, path string) ([]string, error) {
	var list []string
	return list, filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if less {
			if date > info.ModTime().Unix() {
				return nil
			}
		} else {
			if date < info.ModTime().Unix() {
				return nil
			}
		}
		root = filepath.ToSlash(root)
		if info.IsDir() {
			if matchdir {
				list = append(list, root)
				return nil
			}
			return nil
		}
		list = append(list, root)
		return nil
	})
}

func namewalk(reg *regexp.Regexp, matchdir bool, path string) ([]string, error) {
	var list []string
	return list, filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !reg.MatchString(info.Name()) {
			return nil
		}
		root = filepath.ToSlash(root)
		if info.IsDir() {
			if matchdir {
				list = append(list, root)
				return nil
			}
			return nil
		}
		list = append(list, root)
		return nil
	})
}

type fileInfo []os.FileInfo

func (self fileInfo) Less(i, j int) bool {
	return self[i].ModTime().Unix() > self[j].ModTime().Unix()
}
func (self fileInfo) Len() int {
	return len(self)
}
func (self fileInfo) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func ReadDir(path string) ([]os.FileInfo, error) {
	list, err := readDir(path)
	if err != nil {
		return nil, err
	}
	sort.Sort(fileInfo(list))
	return list, err
}

func readDir(path string) ([]os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return ioutil.ReadDir(path)
	}
	return []os.FileInfo{info}, nil
}
