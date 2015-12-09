package all_ssh

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

var Result chan string = make(chan string, 5)
var W *sync.WaitGroup = new(sync.WaitGroup)

func Run(Con *ssh.Client, cmd string) {
	defer Con.Close()
	defer W.Done()
	s, err := Con.NewSession()
	if err != nil {
		fmt.Printf("%s:新建会话失败.命令未执行.", Con.RemoteAddr())
		return
	}
	fmt.Printf("成功连接:%s\n", Con.RemoteAddr())
	buf, err := s.Output(cmd)
	if err != nil {
		fmt.Printf("%s:命令执行失败.", Con.RemoteAddr())
		return
	}
	Result <- fmt.Sprintf("%s 的执行结果:\n%s\n", Con.RemoteAddr(), string(buf))
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
