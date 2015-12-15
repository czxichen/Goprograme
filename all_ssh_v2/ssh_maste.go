package main

import (
	"all_ssh"
	"github.com/fatih/color"
	"os"
	"os/signal"
	"strings"
	"time"
)

var File *os.File

func init() {
	all_ssh.Flag()
	if err := all_ssh.Parse(all_ssh.ArgsInfo.Serverconfig); err != nil {
		os.Exit(1)
	}
	if all_ssh.ArgsInfo.Key != "" {
		list := strings.Split(all_ssh.ArgsInfo.Key, ",")
		all_ssh.ReadKey(list)
	}
	if all_ssh.ArgsInfo.Cmd != "" {
		var err error
		File, err = os.Create("result.txt")
		if err != nil {
			color.Red("创建输出文件失败:%s\n", err)
		}
	}
}
func main() {
	if all_ssh.ArgsInfo.IP != "" {
		color.Yellow("开始登录:%s\n", all_ssh.ArgsInfo.IP)
		var v all_ssh.ConnetctionInfo
		for _, v = range all_ssh.ServerList {
			if v.IP == all_ssh.ArgsInfo.IP {
				break
			}
		}
		v.IP = all_ssh.ArgsInfo.IP
		client := all_ssh.Connection(v)
		if client == nil {
			return
		}
		err := all_ssh.TtyClient(client)
		if err != nil {
			println(err.Error())
		}
		if len(all_ssh.ErrorList) >= 1 {
			color.Red(all_ssh.ErrorList[0])
		}
		return
	}
	if all_ssh.ArgsInfo.File != "" {
		copyfile()
		return
	}
	if all_ssh.ArgsInfo.Cmd != "" {
		runcmd()
		return
	}
	color.Blue("使用%s -h查看帮助.\n", os.Args[0])
}

func copyfile() {
	color.Yellow("开始执行文件发送:")
	info, err := os.Lstat(all_ssh.ArgsInfo.File)
	if err != nil || info.IsDir() {
		color.Blue("检查要发送的文件.")
		return
	}
	for _, v := range all_ssh.ServerList {
		go func() {
			client := all_ssh.Connection(v)
			if client != nil {
				all_ssh.CopyFile(client, all_ssh.ArgsInfo.File, all_ssh.ArgsInfo.Dir)
			}
		}()
	}
	var num int
	var Over chan os.Signal = make(chan os.Signal, 1)
	go signal.Notify(Over, os.Interrupt, os.Kill)
	go result(&num, Over)
	<-Over
	color.Yellow("一共有%d条错误.\n", len(all_ssh.ErrorList))
	for _, v := range all_ssh.ErrorList {
		color.Red(v)
	}
	color.Red("收到结果:%d条\n", num)
}

func runcmd() {
	defer File.Close()
	color.Yellow("开始执行命令:%s\n", all_ssh.ArgsInfo.Cmd)
	color.Yellow("成功解析%d条记录.\n", len(all_ssh.ServerList))
	for _, v := range all_ssh.ServerList {
		go func(v all_ssh.ConnetctionInfo) {
			client := all_ssh.Connection(v)
			if client != nil {
				all_ssh.Run(client, all_ssh.ArgsInfo.Cmd)
			}
		}(v)
	}
	var num int
	var Over chan os.Signal = make(chan os.Signal, 1)
	go signal.Notify(Over, os.Interrupt, os.Kill)
	go result(&num, Over)
	<-Over
	color.Yellow("一共有%d条错误.\n", len(all_ssh.ErrorList))
	for _, v := range all_ssh.ErrorList {
		color.Red(v)
	}
	color.Red("收到结果:%d条\n", num)
}

func result(num *int, Over chan os.Signal) {
	var T <-chan time.Time
	for {
		select {
		case str := <-all_ssh.Result:
			if File != nil {
				File.WriteString(str)
			}
			*num++
			if *num == len(all_ssh.ServerList) {
				Over <- os.Interrupt
			}
			T = time.After(30e9)
		case <-T:
			if *num != len(all_ssh.ServerList) {
				color.Red("获取结果超时.")
			}
			if *num > len(all_ssh.ServerList)/2 {
				Over <- os.Interrupt
			}
		}
	}
}
