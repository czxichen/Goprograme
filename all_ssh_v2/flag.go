package all_ssh

import "flag"

type Args struct {
	Serverconfig string
	Cmd          string
	IP           string
	File         string
	Dir          string
	Key          string
}

var ArgsInfo Args

func Flag() {
	key := flag.String("k", "", "指定key的路径.-k='key1,key2'")
	serverconfig := flag.String("s", "serverlist", "指定服务器列表文件.")
	cmd := flag.String("c", "", "指定执行的命令.-c='echo Hello World'")
	ip := flag.String("l", "", "以终端登录某个服务器.-l=127.0.0.1")
	file := flag.String("f", "", "发送文件.-f=server.zip")
	dir := flag.String("d", "/tmp", "发送文件到目标路径.")
	flag.Parse()
	ArgsInfo = Args{*serverconfig, *cmd, *ip, *file, *dir, *key}
}
