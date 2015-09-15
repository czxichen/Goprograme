package main

import (
	"bufio"
	"cvt"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type cmd_info struct {
	user   string
	passwd string
	ip     string
	cmds   []string
}

func main() {
	var passwd, cmdpath string
	flag.StringVar(&passwd, "p", "passwd", "指定读取主机列表文件,文件格式：root 123456 127.0.0.1 22")
	flag.StringVar(&cmdpath, "c", "cmdpath", "指定读取命令列表文件,每行一条命令.")
	flag.Parse()
	var result chan string = make(chan string, 1)
	var num chan int32 = make(chan int32, 1)
	var returnnum int32 = 0
	cmds, err := readall(cmdpath)
	if err != nil || len(cmds) <= 0 {
		fmt.Println("解析命令文件出错.")
		return
	}
	File, err := os.Open(passwd)
	if err != nil {
		fmt.Println(err)
		return
	}
	read := bufio.NewReader(File)
	for {
		line, _, err := read.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
			return
		}
		passline := split(cvt.Tostring(line))
		if len(passline) != 4 {
			continue
		}
		info := cmd_info{passline[0], passline[1], fmt.Sprintf("%s:%s", passline[2], passline[3]), cmds}
		atomic.AddInt32(&returnnum, 1)
		go client(info, result, num, &returnnum)
	}
	for {
		select {
		case b := <-result:
			fmt.Println(b)
		case x := <-num:
			if x <= 0 {
				return
			}
		}
	}
}

func client(info cmd_info, result chan string, num chan int32, returnnum *int32) {
	defer func() {
		time.Sleep(1e9)
		num <- atomic.AddInt32(returnnum, -1)
	}()

	config := &ssh.ClientConfig{
		User: info.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(info.passwd),
		},
	}
	var client *ssh.Client
	var err error
	for i := 0; i < 5; i++ {
		client, err = ssh.Dial("tcp", info.ip, config)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Printf("%s 建立连接: %s\n", info.ip, err)
		return
	}
	defer client.Close()
	for _, v := range info.cmds {
		session, err := client.NewSession()
		if err != nil {
			fmt.Println("%s 创建Session出错: %s\n", info.ip, err)
			return
		}
		defer session.Close()
		buf, err := session.Output(v)
		if err != nil {
			fmt.Printf("%s 执行命令：%s出错：%s\n", info.ip, v, err)
			continue
		}
		result <- fmt.Sprintf("remote_PC:%s CMD:%s\n%s", info.ip, v, buf)
	}
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

func readall(cmdpath string) ([]string, error) {
	File, err := os.Open(cmdpath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var cmds []string
	read := bufio.NewReader(File)
	for {
		line, _, err := read.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
			return nil, err
		}
		cmds = append(cmds, string(line))
	}
	return cmds, nil
}
