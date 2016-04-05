package MessageFrameWork

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
)

//默认消息头的长度是6.
const (
	HeadLenght = 6
)

var (
	headLenghtError = errors.New("Read Head faild.")
	connctionPool   sync.Pool
)

type HeadConnection struct {
	lock     *sync.RWMutex
	rwc      net.Conn
	CodeType [2]byte
	MsgLen   int64
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
	MsgLen, _ := binary.Varint(head[2:l])
	hc.lock.Lock()
	hc.CodeType = [2]byte{head[0], head[1]}
	hc.MsgLen = MsgLen
	hc.lock.Unlock()
	return nil
}

//返回EOF错误表示这个消息读取完毕.
func (hc *HeadConnection) Read(p []byte) (int, error) {
	if hc.MsgLen <= 0 {
		err := hc.readHead()
		if err != nil {
			return 0, err
		}
	}
	if int64(len(p)) > hc.MsgLen {
		p = p[0:hc.MsgLen]
	}

	hc.lock.Lock()
	defer hc.lock.Unlock()

	n, err := hc.rwc.Read(p)
	hc.MsgLen -= int64(n)
	if hc.MsgLen <= 0 && err == nil {
		err = io.EOF
	}
	return n, err
}

func (hc *HeadConnection) Write(p []byte) (int, error) {
	return hc.rwc.Write(p)
}

func (hc *HeadConnection) Close() {
	hc.rwc.Close()
	putConnction(hc)
}

func (hc *HeadConnection) RemoteAddr() string {
	return hc.rwc.RemoteAddr().String()
}

func NewConnction(con net.Conn) *HeadConnection {
	c := connctionPool.Get()
	if h, ok := c.(*HeadConnection); ok {
		return h
	}
	return &HeadConnection{lock: new(sync.RWMutex), rwc: con}
}

func putConnction(h *HeadConnection) {
	h.rwc = nil
	connctionPool.Put(h)
}