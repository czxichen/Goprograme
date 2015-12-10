package all_ssh

import (
	"os"
	"strings"

	"sftp"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

func CopyFile(conn *ssh.Client, FileName, DirectoryPath string) bool {
	defer conn.Close()
	defer W.Done()
	if !strings.HasSuffix(DirectoryPath, "/") {
		DirectoryPath = DirectoryPath + "/"
	}
	con, err := sftp.NewClient(conn, sftp.MaxPacket(5e9))
	if err != nil {
		color.Red("%s传输文件新建会话错误: %s\n", conn.RemoteAddr(), err)
		return false
	}
	sFile, _ := os.Open(FileName)
	defer sFile.Close()
	dFile := DirectoryPath + FileName
	color.Yellow("%s 目标路径:%s\n", conn.RemoteAddr(), dFile)
	File, err := con.OpenFile(dFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		color.Red("%s 创建文件%s错误: %s \n", conn.RemoteAddr(), dFile, err)
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
	color.Blue("上传%s到%s成功.\n", FileName, conn.RemoteAddr())
	return true
}
