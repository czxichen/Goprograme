package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	dir     string
	file    string
	oldstr  string
	newstr  string
	Reg     *regexp.Regexp
	Err     error
	test    bool
	readall bool
)

func main() {
	flag.StringVar(&dir, "d", "./", "-d directory")
	flag.StringVar(&file, "f", "", "-f *.ini")
	flag.StringVar(&oldstr, "s", "", "-s oldstr")
	flag.StringVar(&newstr, "r", "", "-r newstr")
	flag.BoolVar(&test, "t", false, "-t true")
	flag.BoolVar(&readall, "a", false, "-a true")
	flag.Parse()

	if file == "" {
		fmt.Printf("Args Error: %s must not null\n", "-f")
		return
	}

	if oldstr == "" || newstr == "" {
		if !test {
			fmt.Printf("Args Error: %s and %s must not null.\n\n", "-s", "-r")
			flag.Usage()
			return
		}
	}

	if strings.Index(file, "*") == 0 {
		file = "." + file
	} else {
		file = "^" + file
	}
	file += "$"
	Reg, Err = regexp.Compile(file)
	if Err != nil {
		fmt.Printf("Regexp string error. Error info: %s\n", Err)
		return
	}
	err := filepath.Walk(dir, walk)
	if err != nil {
		fmt.Println(err)
	}
}

func walk(root string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() || !Reg.MatchString(info.Name()) {
		return nil
	}
	if test {
		fmt.Println(root)
		return nil
	}
	if readall {
		ReadAll(root)
	} else {
		tmp := fmt.Sprintf(".%s.tmp", root)
		if err := ReadLine(root, tmp); err != nil {
			return nil
		}
		os.Remove(root)
		os.Rename(tmp, root)
	}
	return nil
}

func ReadLine(root, tmp string) error {
	File, err := os.Open(root)
	if err != nil {
		fmt.Printf("Open file %s faild.Error Info: %s\n", root, err)
		return err
	}
	defer File.Close()
	tmpFile, err := os.Create(tmp)
	if err != nil {
		fmt.Printf("Create temporary files for %s faild,\n", root)
		return err
	}
	defer tmpFile.Close()
	buf := bufio.NewReader(File)
	var num int = 0
	for {
		line, err := buf.ReadSlice('\n')
		num += bytes.Count(line, []byte(oldstr))
		if err != nil {
			if err.Error() == "EOF" {
				line = bytes.Replace(line, []byte(oldstr), []byte(newstr), -1)
				tmpFile.Write(line)
				break
			}
			fmt.Printf("Read %s error.Error Info: %s\n", root, err)
			return nil
		}
		line = bytes.Replace(line, []byte(oldstr), []byte(newstr), -1)
		tmpFile.Write(line)
	}
	fmt.Printf("File: %s Replace: %d\n", root, num)
	return nil
}

func ReadAll(root string) {
	body, err := ioutil.ReadFile(root)
	if err != nil {
		fmt.Printf("Read file %s faild.Error Info: %s\n", root, err)
		return
	}
	num := bytes.Count(body, []byte(oldstr))
	body = bytes.Replace(body, []byte(oldstr), []byte(newstr), -1)
	File, err := os.Create(root)
	if err != nil {
		fmt.Printf("Create replace file %s faild.Error Info: %s\n", root, err)
		return
	}
	defer File.Close()
	File.Write(body)
	if num > 0 {
		fmt.Printf("File: %s Replace: %d\n", root, num)
	}
}
