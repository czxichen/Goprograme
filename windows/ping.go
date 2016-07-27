package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func Lookup(host string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return "", errors.New("unknown host")
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return addrs[rd.Intn(len(addrs))], nil
}

var Data = []byte("abcdefghijklmnopqrstuvwabcdefghi")

type Reply struct {
	Time  int64
	TTL   uint8
	Error error
}

func main() {
	ping, err := Run("www.google.com", 8, Data)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ping.Close()
	ping.Ping(5)
	fmt.Println(ping.PingCount(6))
}

func MarshalMsg(req int, data []byte) ([]byte, error) {
	xid, xseq := os.Getpid()&0xffff, req
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: xid, Seq: xseq,
			Data: data,
		},
	}
	return wm.Marshal(nil)
}

type ping struct {
	Addr string
	Conn net.Conn
	Data []byte
}

func (self *ping) Dail() (err error) {
	self.Conn, err = net.Dial("ip4:icmp", self.Addr)
	if err != nil {
		return err
	}
	return nil
}

func (self *ping) SetDeadline(timeout int) error {
	return self.Conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

func (self *ping) Close() error {
	return self.Conn.Close()
}

func (self *ping) Ping(count int) {
	if err := self.Dail(); err != nil {
		fmt.Println("Not found remote host")
		return
	}
	fmt.Println("Start ping from ", self.Conn.LocalAddr())
	self.SetDeadline(10)
	for i := 0; i < count; i++ {
		r := sendPingMsg(self.Conn, self.Data)
		if r.Error != nil {
			if opt, ok := r.Error.(*net.OpError); ok && opt.Timeout() {
				fmt.Printf("From %s reply: TimeOut\n", self.Addr)
				if err := self.Dail(); err != nil {
					fmt.Println("Not found remote host")
					return
				}
			} else {
				fmt.Printf("From %s reply: %s\n", self.Addr, r.Error)
			}
		} else {
			fmt.Printf("From %s reply: time=%d ttl=%d\n", self.Addr, r.Time, r.TTL)
		}
		time.Sleep(1e9)
	}
}

func (self *ping) PingCount(count int) (reply []Reply) {
	if err := self.Dail(); err != nil {
		fmt.Println("Not found remote host")
		return
	}
	self.SetDeadline(10)
	for i := 0; i < count; i++ {
		r := sendPingMsg(self.Conn, self.Data)
		reply = append(reply, r)
		time.Sleep(1e9)
	}
	return
}

func Run(addr string, req int, data []byte) (*ping, error) {
	wb, err := MarshalMsg(req, data)
	if err != nil {
		return nil, err
	}
	addr, err = Lookup(addr)
	if err != nil {
		return nil, err
	}
	return &ping{Data: wb, Addr: addr}, nil
}

func sendPingMsg(c net.Conn, wb []byte) (reply Reply) {
	start := time.Now()

	if _, reply.Error = c.Write(wb); reply.Error != nil {
		return
	}

	rb := make([]byte, 1500)
	var n int
	n, reply.Error = c.Read(rb)
	if reply.Error != nil {
		return
	}

	duration := time.Now().Sub(start)
	ttl := uint8(rb[8])
	rb = func(b []byte) []byte {
		if len(b) < 20 {
			return b
		}
		hdrlen := int(b[0]&0x0f) << 2
		return b[hdrlen:]
	}(rb)
	var rm *icmp.Message
	rm, reply.Error = icmp.ParseMessage(1, rb[:n])
	if reply.Error != nil {
		return
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		t := int64(duration / time.Millisecond)
		reply = Reply{t, ttl, nil}
	case ipv4.ICMPTypeDestinationUnreachable:
		reply.Error = errors.New("Destination Unreachable")
	default:
		reply.Error = fmt.Errorf("Not ICMPTypeEchoReply %v", rm)
	}
	return
}
