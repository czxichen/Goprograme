package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

const (
	HeadLenght = 6
)

var (
	headLenghtError = errors.New("Read Head faild.")
)

func main() {

}

func server(ip string, handle func(con net.Conn)) error {
	lis, err := net.Listen("tcp", ip)
	if err != nil {
		return err
	}
	defer lis.Close()
	for {
		con, err := lis.Accept()
		if err != nil {
			continue
		}
		go handle(con)
	}
}

type HeadConnection struct {
	lock     sync.RWMutex
	rwc      net.Conn
	r        io.Reader
	codeType [2]byte
	msgLen   int64
}

func (hc *HeadConnection) readHead() error {
	head := make([]byte, HeadLenght)
	l, err := hc.rwc.Read(head)
	if err != nil {
		return err
	}
	if l != HeadLenght {
		return headLenghtError
	}
	msgLen, _ := binary.Varint(head[2:l])
	hc.lock.Lock()
	hc.codeType = [2]byte{head[0], head[1]}
	hc.msgLen = msgLen
	hc.r = io.LimitReader(hc.rwc, msgLen)
	hc.lock.Unlock()
	return nil
}

func (hc *HeadConnection) Read(p []byte) (int, error) {
	if hc.msgLen <= 0 {
		err := hc.readHead()
		if err != nil {
			return 0, err
		}
	}
	n, err := hc.r.Read(p)
	if err == io.EOF {
		return 0, nil
	}
	hc.msgLen -= int64(n)
	return n, err
}

func hanle(con net.Conn) {
	defer con.Close()

}
