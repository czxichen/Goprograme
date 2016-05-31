package filepath

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//path表示搜索的路径,FullDir表示是不是递归查询,MatchDir表示是否匹配目录.
type FindFiles struct {
	Path     string `json:path`
	FullDir  bool   `json:fulldir`
	MatchDir bool   `json:matchdir`
}

func (self FindFiles) NewFind(date int64, reg string) ([]string, int64, error) {
	switch {
	case date != 0 && len(reg) > 0:
		return self.DateAndRegexp(date, reg)
	case date != 0:
		return self.DateFindFile(date)
	case len(reg) > 0:
		return self.RegFindFile(reg)
	}
	return nil, 0, errors.New("Unknow args")
}

//date小于等于0的时候表示查找最近这段时间的文件
func (self FindFiles) DateFindFile(date int64) ([]string, int64, error) {
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
	return datewalk(date, less, self.FullDir, self.MatchDir, self.Path)
}

func (self FindFiles) RegFindFile(reg string) ([]string, int64, error) {
	if strings.Index(reg, "*") == 0 {
		reg = "." + reg
	} else {
		reg = "^" + reg
	}
	reg += "$"
	Reg, err := regexp.Compile(reg)
	if err != nil {
		return []string{}, 0, nil
	}
	if self.FullDir {
		return namewalk(Reg, self.MatchDir, self.Path)
	}
	var size int64
	var list []string
	infos, err := readDir(self.Path)
	if err != nil {
		return list, size, nil
	}
	path := filepath.ToSlash(self.Path)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	for _, v := range infos {
		if Reg.MatchString(v.Name()) {
			if v.IsDir() && !self.MatchDir {
				continue
			}
			list = append(list, path+v.Name())
			size += v.Size()
		}
	}
	return list, size, nil
}

func (self FindFiles) DateAndRegexp(date int64, reg string) ([]string, int64, error) {
	var l []string
	list, size, err := self.RegFindFile(reg)
	if err != nil {
		return l, size, err
	}
	date = date * 24 * 60 * 60
	var less bool = false
	if date <= 0 {
		date = time.Now().Unix() + date
		less = true
	} else {
		date = time.Now().Unix() - date
	}
	for _, v := range list {
		info, err := os.Stat(v)
		if err != nil {
			continue
		}
		if less {
			if date > info.ModTime().Unix() {
				size -= info.Size()
				continue
			}
		} else {
			if date < info.ModTime().Unix() {
				size -= info.Size()
				continue
			}
		}
		l = append(l, v)
	}
	return l, size, nil
}

func datewalk(date int64, less bool, fulldir, matchdir bool, path string) ([]string, int64, error) {
	var list []string
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	var size int64
	if !fulldir {
		infos, err := readDir(path)
		if err != nil {
			return list, size, err
		}
		for _, info := range infos {
			file, ok := dResolve(date, less, matchdir, path, info)
			if ok {
				file = path + file
				list = append(list, file)
				size += info.Size()
			}
		}
		return list, size, nil
	}
	err := filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		_, ok := dResolve(date, less, matchdir, root, info)
		if ok {
			root = filepath.ToSlash(root)
			list = append(list, root)
			size += info.Size()
		}
		return nil
	})
	return list, size, err
}

func dResolve(date int64, less, matchdir bool, root string, info os.FileInfo) (string, bool) {
	if less {
		if date > info.ModTime().Unix() {
			return "", false
		}
	} else {
		if date < info.ModTime().Unix() {
			return "", false
		}
	}
	root = filepath.ToSlash(root)
	if info.IsDir() && !matchdir {
		return "", false
	}

	return info.Name(), true
}

func namewalk(reg *regexp.Regexp, matchdir bool, path string) ([]string, int64, error) {
	var list []string
	var size int64
	err := filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !reg.MatchString(info.Name()) {
			return nil
		}
		root = filepath.ToSlash(root)
		if info.IsDir() && !matchdir {
			return nil
		}
		list = append(list, root)
		size += info.Size()
		return nil
	})
	return list, size, err
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

/*
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
*/
