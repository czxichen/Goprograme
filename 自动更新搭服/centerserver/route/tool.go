package route

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type updateResult struct {
	Path  string   `json:path`
	Files []string `json:files`
}

func TimeStrToInt64(str string) float64 {
	t, err := time.Parse("20060102", str)
	if err != nil {
		return -1
	}
	return float64(t.Unix())
}

func getDirs(path string) []string {
	var list []string
	infolist, err := ioutil.ReadDir(path)
	if err != nil {
		Log.PrintfE("读取目录失败:%s %s\n", path, err)
		return list
	}
	for _, v := range infolist {
		if v.IsDir() {
			list = append(list, v.Name())
		}
	}
	return list
}

func getNewDir(path string) string {
	//找出日期最大的的一个目录.
	l := getDirs(path)
	var unixtimelist []float64
	for _, i := range l {
		//		fmt.Println(i)
		if t := TimeStrToInt64(i); t != -1 {
			unixtimelist = append(unixtimelist, t)
		}
	}
	if len(unixtimelist) <= 0 {
		return ""
	}
	sort.Float64s(unixtimelist)
	recentlytime := int64(unixtimelist[len(unixtimelist)-1])
	return time.Unix(int64(recentlytime), 0).Format("20060102")
}

func parseUpdatedir(dirpath string) Namesort {
	dirname := filepath.Base(dirpath)
	var l []string
	list, err := ioutil.ReadDir(dirpath)
	if err != nil {
		Log.PrintfE("%s\n", err)
		return l
	}
	for _, v := range list {
		if strings.Contains(v.Name(), dirname) {
			if toint(v.Name()) == -1 {
				continue
			}
			l = append(l, v.Name())
		}
	}
	return l
}

type Namesort []string

func (self Namesort) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self Namesort) Len() int {
	return len(self)
}

func (self Namesort) Less(i, j int) bool {
	ai := toint(self[i])
	aj := toint(self[j])
	if ai != -1 && aj != -1 {
		return ai < aj
	}
	return false
}

func toint(str string) int {
	str = strings.Split(str, ".")[0]
	list := strings.Split(str, "_")
	if len(list) <= 1 {
		return -1
	}
	n := list[len(list)-1]
	num, err := strconv.Atoi(n)
	if err != nil {
		return -1
	}
	return num
}

type info []os.FileInfo

func (self info) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self info) Less(i, j int) bool {
	return self[i].ModTime().Unix() > self[j].ModTime().Unix()
}

func (self info) Len() int {
	return len(self)
}

func (self info) Sort() {
	sort.Sort(self)
}
