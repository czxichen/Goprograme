package all_ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type ConnetctionInfo struct {
	User     string
	Passwd   string
	IP       string
	Port     string
	ExecUser string
}

var ErrorList []string
var privateKey []ssh.AuthMethod

func Connection(info ConnetctionInfo) *ssh.Client {
	var auths []ssh.AuthMethod
	auths = append(auths, privateKey...)
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}
	auths = append(auths, ssh.Password(info.Passwd))
	config := ssh.ClientConfig{
		User: info.User,
		Auth: auths,
	}
	for i := 0; i < 3; i++ {
		//conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", info.IP, info.Port), &config)
		conn, err := ssh.DialTimeOut("tcp", fmt.Sprintf("%s:%s", info.IP, info.Port), 30, &config)
		if err == nil {
			return conn
		}
		if i == 2 && err != nil {
			ErrorList = append(ErrorList, fmt.Sprintf("连接%s失败:%s\n", info.IP, err))
			return nil
		}
		time.Sleep(1e9)
	}
	return nil
}
