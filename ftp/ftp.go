package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/jlaffaye/ftp"
)

var (
	files     int = 2
	dstdir    string
	sourcedir = "C:/SnailGame/DatabaseBackup/Transaction/"
	dirs      = []string{"", "chargeDB", "sub5TransloadBak"}
)

func main() {
	client, err := getFtpConn("ip", "user", "passwd")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Quit()
	FilesInfo := getFiles(sourcedir, files)
	for _, info := range FilesInfo {
		flushdstdir(client, info)
		sendFile(client, sourcedir+info.Name(), dstdir+info.Name())
	}
}

func flushdstdir(client *ftp.ServerConn, info os.FileInfo) {
	dirs[0] = info.ModTime().Format("20060102")
	var path string
	for _, p := range dirs {
		path = path + "/" + p
		err := client.ChangeDir(path)
		if err != nil {
			err = client.MakeDir(path)
			if err != nil {
				fmt.Println("Create dir faild:", err)
				return
			}
		}
	}
	if dstdir != path+"/" {
		dstdir = path + "/"
	}
}

func getFiles(path string, num int) []os.FileInfo {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	sort.Sort(fileinfo(infos))
	if len(infos) < num {
		return infos
	}
	return infos[:num]
}

func sendFile(client *ftp.ServerConn, src, dsc string) error {
	list, err := client.NameList(dsc)
	if err != nil {
		return fmt.Errorf("Upload file faild,Error info:%s", err.Error())
	}
	if len(list) > 0 {
		return fmt.Errorf("Upload file faild,Error info:%s", "Remote is exist")
	}
	File, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Open local file faild,Error info:%s", err.Error())
	}
	defer File.Close()
	return client.Stor(dsc, File)
}

func getFtpConn(addr, user, passwd string) (*ftp.ServerConn, error) {
	f, err := ftp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("Connection faild,Error info:%s", err.Error())
	}
	err = f.Login(user, passwd)
	if err != nil {
		return nil, fmt.Errorf("Login faild,Error info:%s", err.Error())
	}
	return f, nil
}

type fileinfo []os.FileInfo

func (self fileinfo) Len() int {
	return len(self)
}

func (self fileinfo) Less(i, j int) bool {
	return self[i].ModTime().Unix() > self[j].ModTime().Unix()
}

func (self fileinfo) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
