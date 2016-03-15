package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	config := flag.String("f", "replations.ini", "-f replations.ini")
	dir := flag.String("d", "config", "-d configdir")
	flag.Parse()
	M, err := ParseConfig(*config)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = AutoReplace(*dir, M)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ParseConfig(configpath string) (map[string]string, error) {
	file, err := os.Open(configpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	M := make(map[string]string)
	buf := bufio.NewReader(file)
	var num int
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		list := bytes.Split(line, []byte("="))
		if len(list) != 2 {
			return nil, fmt.Errorf("第%d行 出现多次'='", num)
		}
		M[string(bytes.TrimSpace(list[0]))] = string(bytes.TrimSpace(list[1]))
		num++
	}
	return M, nil
}

func AutoReplace(dirpath string, variablesMap map[string]string) error {
	Files, err := getFiles(dirpath)
	if err != nil {
		return err
	}
	for fileName, body := range Files {
		for k, v := range variablesMap {
			body = bytes.Replace(body, []byte(k), []byte(fmt.Sprintf("{{.%s}}", v)), -1)
		}
		F, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			F.Close()
			return err
		}
		F.Write(body)
		F.Close()
	}
	return nil
}

func getFiles(path string) (map[string][]byte, error) {
	if !bytes.HasSuffix([]byte(path), []byte("/")) {
		path = path + "/"
	}
	files, err := ioutil.ReadDir(path)
	if err != nil || len(files) <= 0 {
		return nil, err
	}
	var fileInfo map[string][]byte = make(map[string][]byte)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		F, err := os.Open(path + file.Name())
		if err != nil {
			return nil, err
		}
		buf := make([]byte, file.Size())
		n, err := io.ReadFull(F, buf)
		if err != nil {
			return nil, err
		}
		if bytes.Contains(buf[:n], []byte{0}) {
			continue
		}
		fileInfo[path+file.Name()] = buf[:n]
	}
	return fileInfo, nil
}
