package all_ssh

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

var Result chan string = make(chan string, 10)

func Run(Con *ssh.Client, cmd string) {
	defer Con.Close()
	s, err := Con.NewSession()
	if err != nil {
		color.Red("%s:新建会话失败.命令未执行.", Con.RemoteAddr())
		return
	}
	fmt.Printf("成功连接:%s\n", Con.RemoteAddr())
	buf, err := s.Output(cmd)
	if err != nil {
		color.Red("%s:命令执行失败.", Con.RemoteAddr())
		return
	}
	str := fmt.Sprintf("%s 的执行结果:\n%s\n", Con.RemoteAddr().String(), string(buf))
	fmt.Println(str)
	Result <- str
}

func TtyClient(Con *ssh.Client) error {
	defer Con.Close()
	session, err := Con.NewSession()
	if err != nil {
		return err
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm", 25, 100, modes)
	if err != nil {
		return err
	}
	session.Shell()
	return session.Wait()
}
