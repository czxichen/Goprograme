package all_ssh

import (
	"bufio"
	"github.com/fatih/color"
	"os"
	"strings"
)

var ServerList []ConnetctionInfo

func Parse(path string) error {
	File, err := os.Open(path)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(File)
	var num int = 1
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				return err
			}
			break
		}
		list := split(string(line))
		if len(list) != 5 {
			color.Red("ErrorLine:%d %s\n", num, string(line))
			continue
		}
		info := ConnetctionInfo{list[0], list[1], list[2], list[3], list[4]}
		ServerList = append(ServerList, info)
	}
	return nil
}

func split(str string) []string {
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
