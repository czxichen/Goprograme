package main

import (
	"bufio"
	"fmt"
	"os"
)

var File, dFile *os.File
var err error

func main() {
	if len(os.Args) != 3 {
		fmt.Println("检查参数.")
		fmt.Printf("用法:\n%s copy or join filepath\n", os.Args[0])
		return
	}
	Init(os.Args[2])
	switch os.Args[1] {
	case "copy":
		Copy()
	case "join":
		Join()
	default:
		fmt.Println("此参数未定义")
	}
}

func Init(filepath string) {
	File, err = os.Open(filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dFile, err = os.Create("result.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Copy() {
	buf := bufio.NewReader(File)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println(err)
			}
			File.Close()
			dFile.Close()
			os.Exit(1)
		}
		dFile.Write(line)
		dFile.Write([]byte("\r\n"))
		dFile.Write(line)
		dFile.Write([]byte("\r\n"))
	}
}
func Join() {
	var ok bool = true
	buf := bufio.NewReader(File)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println(err)
			}
			File.Close()
			dFile.Close()
			os.Exit(1)
		}
		dFile.Write(line)
		if ok {
			ok = false
			dFile.Write([]byte("\t"))
		} else {
			ok = true
			dFile.Write([]byte("\r\n"))
		}
	}
}
