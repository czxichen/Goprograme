package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/czxichen/AutoWork/tools/split"
	"golang.org/x/crypto/ssh"
)

var (
	passwd  = flag.String("p", "", "-p passwd ָ������.")
	user    = flag.String("u", "root", "-u root ָ����¼�û�.")
	cfg     = flag.String("c", "serverlist", "-c serverlist ָ��serverlist")
	ip_port = flag.String("i", "", "-i ip:port ָ��Ŀ�������IP�˿�,�����-p���ʹ�÷�����Ч.")
	dpath   = flag.String("d", "", "-d /tmp/20160531.zip ָ�����͵���·��,����Ϊ��.")
	spath   = flag.String("s", "", "-s 20160531.zip ָ��Ҫ�����ļ���·��,����Ϊ��.")
)

func main() {
	flag.Parse()

	if *dpath == "" || *spath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	File, err := os.Open(*spath)
	if err != nil {
		fmt.Println("���ļ�ʧ��:", err)
		os.Exit(1)
	}
	info, _ := File.Stat()
	defer File.Close()

	if *ip_port != "" && *passwd != "" {
		Client, err := dail(*user, *passwd, *ip_port)
		if err != nil {
			fmt.Printf("����%sʧ��.\n", err)
			os.Exit(1)
		}
		scp(Client, File, info.Size(), *dpath)
		return
	}
	var list [][]string
	ok := (*passwd != "" && *ip_port == "")
	list = config(*cfg, ok)
	if len(list) <= 0 {
		fmt.Println("serverlist ����Ϊ��.")
		os.Exit(1)
	}
	for _, v := range list {
		if ok {
			*ip_port = v[0]
		} else {
			*user = v[0]
			*passwd = v[1]
			*ip_port = v[2]
		}
		Client, err := dail(*user, *passwd, *ip_port)
		if err != nil {
			fmt.Printf("����%sʧ��.\n", err)
			continue
		}
		scp(Client, File, info.Size(), *dpath)
	}
}

func dail(user, password, ip_port string) (*ssh.Client, error) {
	PassWd := []ssh.AuthMethod{ssh.Password(password)}
	Conf := ssh.ClientConfig{User: user, Auth: PassWd}
	return ssh.Dial("tcp", ip_port, &Conf)
}

func scp(Client *ssh.Client, File io.Reader, size int64, path string) {
	filename := filepath.Base(path)
	dirname := strings.Replace(filepath.Dir(path), "\\", "/", -1)
	defer Client.Close()

	session, err := Client.NewSession()
	if err != nil {
		fmt.Println("����Sessionʧ��:", err)
		return
	}
	go func() {
		w, _ := session.StdinPipe()
		fmt.Fprintln(w, "C0644", size, filename)
		io.CopyN(w, File, size)
		fmt.Fprint(w, "\x00")
		w.Close()
	}()

	if err := session.Run(fmt.Sprintf("/usr/bin/scp -qrt %s", dirname)); err != nil {
		fmt.Println("ִ��scp����ʧ��:", err)
		session.Close()
		return
	} else {
		fmt.Printf("%s ���ͳɹ�.\n", Client.RemoteAddr())
		session.Close()
	}

	if session, err = Client.NewSession(); err == nil {
		defer session.Close()
		buf, err := session.Output(fmt.Sprintf("/usr/bin/md5sum %s", path))
		if err != nil {
			fmt.Println("���md5ʧ��:", err)
			return
		}
		fmt.Printf("%s ��MD5:\n%s\n", Client.RemoteAddr(), string(buf))
	}
}

func config(path string, ok bool) (list [][]string) {
	File, err := os.Open(path)
	if err != nil {
		fmt.Printf("�������ļ�ʧ��:%s\n", err)
		os.Exit(1)
	}
	defer File.Close()
	buf := bufio.NewReader(File)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		str := strings.TrimSpace(string(line))
		strs := split.Split(str)
		if ok {
			if len(strs) != 1 {
				continue
			}
		} else {
			if len(strs) != 3 {
				continue
			}
		}
		list = append(list, strs)
	}
	return
}
