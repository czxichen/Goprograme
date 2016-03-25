package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ConnetctionInfo struct {
	User   string
	Passwd string
	IP     string
}

var (
	src, dst string
	debug    bool
)

func main() {
	var info ConnetctionInfo
	flag.StringVar(&info.User, "u", "root", "-u root")
	flag.StringVar(&info.Passwd, "p", "", "-p 123456")
	flag.StringVar(&info.IP, "i", "", "-i 127.0.0.1:22")
	flag.StringVar(&src, "s", "", "-s filename")
	flag.StringVar(&dst, "d", "", "-d dfile")
	flag.BoolVar(&debug, "-debug", false, "-d true")
	flag.Parse()
	client, err := con(info.User, info.IP, info.Passwd)
	if err != nil {
		if debug {
			fmt.Println(err)
		}
		os.Exit(-1)
	}
	if !CopyFile(client, src, dst) {
		os.Exit(-1)
	}
}

func con(user, ip, passwd string) (*ssh.Client, error) {
	return ssh.Dial("tcp", ip, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(passwd)},
	})
}

func CopyFile(conn *ssh.Client, FileName, DirectoryPath string) bool {
	defer conn.Close()
	if !strings.HasSuffix(DirectoryPath, "/") {
		DirectoryPath = DirectoryPath + "/"
	}
	con, err := sftp.NewClient(conn, sftp.MaxPacket(5e9))
	if err != nil {
		if debug {
			fmt.Println(err)
		}
		return false
	}
	sFile, _ := os.Open(FileName)
	defer sFile.Close()
	dFile := DirectoryPath + FileName
	File, err := con.OpenFile(dFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		if debug {
			fmt.Println(err)
		}
		return false
	}
	defer File.Close()
	for {
		buf := make([]byte, 1024)
		n, err := sFile.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return false
		}
		File.Write(buf[:n])
	}
	return true
}
