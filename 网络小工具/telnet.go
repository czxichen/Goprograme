package main

import (
	"flag"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ip      string
	timeout int
	action  string
)

func init() {
	flag.IntVar(&timeout, "t", 10, "-t 10 指定超时时间,单位:s")
	flag.StringVar(&action, "a", "", "-a 'w_Hello','r_10' 指定如何交互,w_'body'表示发.body表示发送的内容,r_'length'表示收.length表示收的长度")
	flag.StringVar(&ip, "h", "", "-h host:port 指定远程的IP和端口")
	flag.Parse()
	if ip == "" {
		println("IP不能为空")
		os.Exit(1)
	}
}

func main() {
	if action == "" {
		if PortIsOpen(ip, timeout) {
			println("连接成功")
		} else {
			println("连接失败")
		}
		os.Exit(1)
	}
	actionlist := strings.Split(action, ",")
	Telnet(actionlist, ip, timeout)
}

func PortIsOpen(ip string, timeout int) bool {
	con, err := net.DialTimeout("tcp", ip, time.Duration(timeout)*time.Second)
	if err != nil {
		println("Error Info:", err.Error())
		return false
	}
	con.Close()
	return true
}

func Telnet(action []string, ip string, timeout int) {
	con, err := net.DialTimeout("tcp", ip, time.Duration(timeout)*time.Second)
	if err != nil {
		return
	}
	defer con.Close()
	con.SetReadDeadline(time.Now().Add(time.Second * 5))
	for _, v := range action {
		v = strings.TrimSpace(v)
		l := strings.SplitN(v, "_", 2)
		if len(l) < 2 {
			return
		}
		switch l[0] {
		case "r":
			var n int
			n, err = strconv.Atoi(l[1])
			if err != nil {
				println("转换失败:", err.Error())
				return
			}
			p := make([]byte, n)
			n, err = con.Read(p)
			if err != nil {
				println("读取错误:", err.Error())
				return
			}
			println(string(p[:n]))
		case "w":
			msg := []byte(l[1])
			msg = append(msg, []byte("\r\n")...)
			_, err = con.Write(msg)
			if err != nil {
				println("Send Error:", err.Error())
				return
			}
		}
	}
	return
}
