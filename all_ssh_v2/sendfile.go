package all_ssh

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func CopyFile(conn *ssh.Client, FileName, DirectoryPath string) bool {
	if !strings.HasSuffix(DirectoryPath, "/") {
		DirectoryPath = DirectoryPath + "/"
	}
	con, err := sftp.NewClient(conn, sftp.MaxPacket(5e9))
	if err != nil {
		fmt.Printf("%s传输文件新建会话错误: %s\n", conn.RemoteAddr(), err)
		return false
	}
	sFile, _ := os.Open(FileName)
	defer sFile.Close()
	dFile := fmt.Sprintf("%s%s", DirectoryPath, FileName)
	File, err := con.OpenFile(dFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		fmt.Printf("%s 创建文件%s错误: %s \n", conn.RemoteAddr(), dFile, err)
		return false
	}
	defer File.Close()
	io.Copy(File, sFile)
	fmt.Printf("上传%s到%s成功.\n", FileName, conn.RemoteAddr())
	return true
}
