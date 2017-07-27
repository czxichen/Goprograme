package proxy

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type remote struct {
	mu       sync.Mutex
	addr     string
	inactive bool
}

func (r *remote) inactivate() {
	r.mu.Lock()
	r.inactive = true
	r.mu.Unlock()
}

//尝试后端服务是否可用
func (r *remote) tryReactivate() error {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}
	conn.Close()
	r.mu.Lock()
	r.inactive = false
	r.mu.Unlock()
	return nil
}

func (r *remote) isActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return !r.inactive
}

type TCPProxy struct {
	Listener        net.Listener
	Endpoints       []string
	MonitorInterval time.Duration

	donec chan struct{}

	mu         sync.Mutex
	remotes    []*remote
	nextRemote int
}

func (tp *TCPProxy) Run() error {
	tp.donec = make(chan struct{})
	if tp.MonitorInterval == 0 {
		tp.MonitorInterval = 5 * time.Minute
	}

	for _, ep := range tp.Endpoints {
		tp.remotes = append(tp.remotes, &remote{addr: ep})
	}

	log.Printf("ready to proxy client requests to %s", strings.Join(tp.Endpoints, ","))

	go tp.runMonitor()
	for {
		in, err := tp.Listener.Accept()
		if err != nil {
			return err
		}
		go tp.serve(in)
	}
}

func (tp *TCPProxy) numRemotes() int {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return len(tp.remotes)
}

func (tp *TCPProxy) serve(in net.Conn) {
	var (
		err error
		out net.Conn
	)

	for i := 0; i < tp.numRemotes(); i++ {
		remote := tp.pick()
		if !remote.isActive() {
			continue
		}
		out, err = net.Dial("tcp", remote.addr)
		if err == nil {
			break
		}
		//连接成功则修改状态可用
		remote.inactivate()
		log.Printf("deactivated endpoint [%s] due to %v for %v", remote.addr, err, tp.MonitorInterval)
	}

	if out == nil {
		in.Close()
		return
	}

	go func() {
		io.Copy(in, out)
		in.Close()
		out.Close()
	}()

	io.Copy(out, in)
	out.Close()
	in.Close()
}

//轮询的机制选取代理的后端
func (tp *TCPProxy) pick() (picked *remote) {
	tp.mu.Lock()
	picked = tp.remotes[tp.nextRemote]
	tp.nextRemote = (tp.nextRemote + 1) % len(tp.remotes)
	tp.mu.Unlock()
	return picked
}

//定时监空后端是否可用.
func (tp *TCPProxy) runMonitor() {
	for {
		select {
		case <-time.After(tp.MonitorInterval):
			tp.mu.Lock()
			for _, r := range tp.remotes {
				if !r.isActive() {
					go func() {
						if err := r.tryReactivate(); err != nil {
							log.Printf("failed to activate endpoint [%s] due to %v (stay inactive for another %v)", r.addr, err, tp.MonitorInterval)
						} else {
							log.Printf("activated %s", r.addr)
						}
					}()
				}
			}
			tp.mu.Unlock()
		case <-tp.donec:
			return
		}
	}
}

//关闭服务.
func (tp *TCPProxy) Stop() {
	tp.Listener.Close()
	close(tp.donec)
}
