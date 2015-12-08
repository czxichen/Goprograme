package tool

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

var Temp_Path map[string]string = make(map[string]string)
var ConfigTemplate []string

func GetPathConfig(path string) {
	File, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	defer File.Close()
	M := make(map[string]string)
	Buf := bufio.NewReader(File)
	var linenum int = 1
	for {
		line, _, err := Buf.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
			os.Exit(5)
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		list := bytes.Split(line, []byte("="))
		if len(list) != 2 {
			fmt.Printf("check config %s ,line %d\n", path, linenum)
			os.Exit(6)
		}
		key := string(bytes.TrimSpace(list[0]))
		value := string(bytes.TrimSpace(list[1]))
		TestPath(key)
		M[key] = value
		linenum++
	}
	if len(M) < 1 {
		fmt.Printf("config %s can't emptey!\n", path)
		os.Exit(7)
	}
	for k, _ := range M {
		ConfigTemplate = append(ConfigTemplate, k)
	}
	Temp_Path = M
}
