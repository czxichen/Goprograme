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

//CodeType是读取消息头的前两个字节,用来标记消息.
//MsgLen表示消息的剩余长度.
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
//返回EOF错误要判断下消息是否为空.
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

//关闭conection链接,并还回池子.
func (hc *HeadConnection) Close() {
	hc.rwc.Close()
	putConnction(hc)
}

func (hc *HeadConnection) RemoteAddr() string {
	return hc.rwc.RemoteAddr().String()
}

//返回一个HeadConnection指针.
func NewConnction(con net.Conn) *HeadConnection {
	c := connctionPool.Get()
	if h, ok := c.(*HeadConnection); ok {
		h.rwc = con
		return h
	}
	return &HeadConnection{lock: new(sync.RWMutex), rwc: con}
}

func putConnction(h *HeadConnection) {
	h.rwc = nil
	connctionPool.Put(h)
}

//4个字节的最大消息长度134217727.
//如果消息长度超过4个字节可以改变HeadLenght的值.
//首先判断下消息长度有没有超出限制.
//head的前两个字节是设置一下消息类型.
func NewHeadByte(t [2]byte, l int64) []byte {
	if (HeadLenght-2)*7-1 < l {
		return nil
	}
	b := make([]byte, HeadLenght)
	b[0], b[1] = t[0], t[1]
	binary.PutVarint(b[2:HeadLenght], l)
	return b
}
