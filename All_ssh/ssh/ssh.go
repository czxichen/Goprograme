package ssh

import (
	"fmt"
	"net"
	"os"
	"sftp"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type RunInfo struct {
	User     string
	Passwd   string
	Ip       string
	RunUser  string
	FileName string
}

var Result chan string = make(chan string, 1)
var ErrorList []string

func Connection(info RunInfo) *ssh.Client {
	var auths []ssh.AuthMethod
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}
	auths = append(auths, ssh.Password(info.Passwd))
	config := ssh.ClientConfig{
		User: info.User,
		Auth: auths,
	}
	for i := 0; i < 5; i++ {
		conn, err := ssh.Dial("tcp", info.Ip, &config)
		if err == nil {
			return conn
		}
		if i == 4 && err != nil {
			ErrorList = append(ErrorList, fmt.Sprintf("连接%s失败:%s\n", info.Ip, err))
			return nil
		}
		time.Sleep(1e9)
	}
	return nil
}
func Exec(con *ssh.Client, cmd string) (string, error) {
	session, err := con.NewSession()
	if err != nil {
		return "", err
	}
	buf, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func CopyFile(conn *ssh.Client, FileName string, body []byte) bool {
	fmt.Printf("开始向%s发送脚本.\n", conn.RemoteAddr())
	con, err := sftp.NewClient(conn, sftp.MaxPacket(5e9))
	if err != nil {
		fmt.Printf("%s新建会话错误: %s\n", conn.RemoteAddr(), err)
		return false
	}
	con.Mkdir("/tmp/.cache")
	File, err := con.OpenFile(FileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		fmt.Printf("%s创建文件错误: %s \n", conn.RemoteAddr(), err)
		return false
	}
	File.Write(body)
	File.Chmod(0777)
	File.Close()
	return true
}
func Run(con *ssh.Client, FileName string, buf []byte, Lock *sync.WaitGroup) {
	defer Lock.Done()
	defer con.Close()
	if CopyFile(con, FileName, buf) {
		fmt.Printf("%s开始执行脚本.\n", con.RemoteAddr())
		str, err := Exec(con, fmt.Sprintf("su - root %s", FileName))
		if err != nil {
			fmt.Printf("%s执行出错:%s\n", con.RemoteAddr(), err)
			return
		}
		Result <- fmt.Sprintf("%s的执行结果:%s", con.RemoteAddr(), str)
	}
}
