package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"ssh"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("参数错误.")
		return
	}
	server_list := os.Args[1]
	FileName := os.Args[2]
	List := Parse(server_list)
	if len(List) < 1 {
		fmt.Printf("解析配置%s出错\n", server_list)
		return
	}
	fmt.Printf("配置服务端列表解析完成,共%d条.\n", len(List))
	buf, err := ioutil.ReadFile(FileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	File, err := os.Create("result.log")
	if err != nil {
		fmt.Println("建立日志文件失败.", err)
		return
	}
	Lock := new(sync.WaitGroup)
	defer File.Close()
	for _, v := range List {
		Lock.Add(1)
		go func(Lock *sync.WaitGroup, Runinfo ssh.RunInfo) {
			con := ssh.Connection(Runinfo)
			if con == nil {
				Lock.Done()
				return
			}
			fmt.Printf("%s连接成功.\n", con.RemoteAddr())
			go ssh.Run(con, fmt.Sprintf("/tmp/.cache/%s", FileName), buf, Lock)
		}(Lock, v)
	}
	var num int
	go func() {
		for {
			select {
			case str := <-ssh.Result:
				File.WriteString(fmt.Sprint(str))
				num++
			}
		}
	}()
	Lock.Wait()
	time.Sleep(1e9)
	for _, v := range ssh.ErrorList {
		fmt.Println(v)
	}
	fmt.Printf("共计接受结果: %d条", num)
}

func Parse(FileName string) []ssh.RunInfo {
	var ListRunInfo []ssh.RunInfo
	buf, err := os.Open(FileName)
	if err != nil {
		fmt.Println("打开服务端列表失败: ", err)
		return ListRunInfo
	}
	defer buf.Close()
	fmt.Printf("开始解析:%s\n", FileName)
	R := bufio.NewReader(buf)
	for {
		line, _, err := R.ReadLine()
		if err != nil {
			break
		}
		list := split(string(line))
		if len(list) != 4 {
			continue
		}
		ListRunInfo = append(ListRunInfo, ssh.RunInfo{list[0], list[1], list[2], list[3], FileName})
	}
	return ListRunInfo
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
