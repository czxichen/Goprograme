package tool

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func Walkconfig(Path string) {
	filepath.Walk(Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		Check(path)
		return nil
	})
}

func Check(path string) {
	File, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer File.Close()
	buf := bufio.NewReader(File)
	var num int = 1
	var errornum int = 0
	s := []byte("{{")
	e := []byte("}}")
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return
		}
		if bytes.Count(line, s) != bytes.Count(line, e) {
			fmt.Printf("Line%d: %s\n", num, string(line))
			errornum++
			continue
		}
		if bytes.Count(line, []byte("{{.}}")) != 0 {
			fmt.Printf("Line%d: %s\n", num, string(line))
			errornum++
			continue
		}

		for i := 0; i < bytes.Count(line, s); i++ {
			first := bytes.Index(line, s)
			last := bytes.Index(line, e)
			if first == -1 || last == -1 {
				continue
			}
			if bytes.Index(line[first:last], []byte("{{.")) != 0 {
				fmt.Printf("Error Line %d: %s\n", num, string(line))
				errornum++
				break
			}
			line = line[last:]
		}
	}
	if errornum != 0 {
		fmt.Printf("Error num %d From %s\n", errornum, path)
		return
	}
	return
}
