package filepath

import (
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

//date小于等于0的时候表示查找最近这段时间的文件
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
	return datewalk(date, less, self.FullDir, self.MatchDir, self.Path)
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
			if v.IsDir() && !self.MatchDir {
				continue
			}
			list = append(list, path+v.Name())
		}
	}
	return list, nil
}

func (self FindFiles) DateAndRegexp(date int64, reg string) ([]string, error) {
	var l []string
	list, err := self.RegFindFile(reg)
	if err != nil {
		return l, err
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
				continue
			}
		} else {
			if date < info.ModTime().Unix() {
				continue
			}
		}
		l = append(l, v)
	}
	return l, nil
}

func datewalk(date int64, less bool, fulldir, matchdir bool, path string) ([]string, error) {
	var list []string
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	if !fulldir {
		infos, err := readDir(path)
		if err != nil {
			return list, err
		}
		for _, info := range infos {
			file, ok := dResolve(date, less, matchdir, path, info)
			if ok {
				file = path + file
				list = append(list, file)
			}
		}
		return list, nil
	}
	return list, filepath.Walk(path, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		_, ok := dResolve(date, less, matchdir, root, info)
		if ok {
			root = filepath.ToSlash(root)
			list = append(list, root)
		}
		return nil
	})
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
