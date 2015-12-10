package main

import (
	"all_ssh"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func init() {
	all_ssh.Flag()
	if err := all_ssh.Parse(all_ssh.ArgsInfo.Serverconfig); err != nil {
		os.Exit(1)
	}
	if all_ssh.ArgsInfo.Key != "" {
		list := strings.Split(all_ssh.ArgsInfo.Key, ",")
		all_ssh.ReadKey(list)
	}
}

func main() {
	if all_ssh.ArgsInfo.IP != "" {
		color.Yellow("开始登录:%s\n", all_ssh.ArgsInfo.IP)
		for _, v := range all_ssh.ServerList {
			if v.IP == all_ssh.ArgsInfo.IP {
				client := all_ssh.Connection(v)
				if client == nil {
					return
				}
				err := all_ssh.TtyClient(client)
				if err != nil {
					println(err.Error())
				}
				return
			}
		}
		return
	}
	if all_ssh.ArgsInfo.File != "" {
		color.Yellow("开始执行文件发送:")
		info, err := os.Lstat(all_ssh.ArgsInfo.File)
		if err != nil || info.IsDir() {
			color.Blue("检查要发送的文件.")
			return
		}
		for _, v := range all_ssh.ServerList {
			client := all_ssh.Connection(v)
			if client != nil {
				all_ssh.W.Add(1)
				go all_ssh.CopyFile(client, all_ssh.ArgsInfo.File, all_ssh.ArgsInfo.Dir)
			}
		}
		all_ssh.W.Wait()
		return
	}
	if all_ssh.ArgsInfo.Cmd != "" {
		color.Yellow("开始执行命令:%s\n", all_ssh.ArgsInfo.Cmd)
		for _, v := range all_ssh.ServerList {
			client := all_ssh.Connection(v)
			if client != nil {
				all_ssh.W.Add(1)
				go all_ssh.Run(client, all_ssh.ArgsInfo.Cmd)
			}
		}
		go func() {
			File, err := os.Create("result.txt")
			if err != nil {
				println(err.Error())
				select {
				case str := <-all_ssh.Result:
					color.Green(str)
				}
			} else {
				defer File.Close()
				for {
					select {
					case str := <-all_ssh.Result:
						File.WriteString(str)
						color.Green(str)
					}
				}
			}
		}()
		all_ssh.W.Wait()
		time.Sleep(1e9)
		color.Yellow("一共有%d条错误.\n", len(all_ssh.ErrorList))
		for _, v := range all_ssh.ErrorList {
			color.Red(v)
		}
		return
	}
	color.Blue("使用%s -h查看帮助.\n", os.Args[0])
}
