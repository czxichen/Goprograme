package main

import (
	"all_ssh"
	"fmt"
	"os"
	"time"
)

func init() {
	all_ssh.Flag()
	if err := all_ssh.Parse(all_ssh.ArgsInfo.Serverconfig); err != nil {
		os.Exit(1)
	}
}

func main() {
	if all_ssh.ArgsInfo.IP != "" {
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
		for _, v := range all_ssh.ServerList {
			client := all_ssh.Connection(v)
			if client != nil {
				all_ssh.CopyFile(client, all_ssh.ArgsInfo.File, all_ssh.ArgsInfo.Dir)
			}
		}
		return
	}
	if all_ssh.ArgsInfo.Cmd != "" {
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
					fmt.Println(str)
				}
			} else {
				select {
				case str := <-all_ssh.Result:
					File.WriteString(str)
				}
			}
		}()
		all_ssh.W.Wait()
		time.Sleep(1e9)
		return
	}
}
