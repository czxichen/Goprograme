package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	client("root", "dijielin", "127.0.0.1:5506")
}

func client(user, passwd, ip string) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
	}
	client, err := ssh.Dial("tcp", ip, config)
	if err != nil {
		fmt.Println("建立连接：", err)
		return
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("创建Session出错：", err)
		return
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Println("创建文件描述符：", err)
		return
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		fmt.Println("获取窗口宽高：", err)
		return
	}
	defer terminal.Restore(fd, oldState)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		fmt.Println("创建终端出错:", err)
		return
	}
	err = session.Shell()
	if err != nil {
		fmt.Println("执行Shell出错:", err)
		return
	}
	err = session.Wait()
	if err != nil {
		fmt.Println("执行Wait出错:", err)
		return
	}
}
