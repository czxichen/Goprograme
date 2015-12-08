package tool

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var HeadList []string
var Replational []string

func Matching(srcip string) string {
	for _, v := range Replational {
		if strings.Contains(v, srcip) {
			return v
		}
	}
	return ""
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
