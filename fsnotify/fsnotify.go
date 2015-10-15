package main

import (
	"flag"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-fsnotify/fsnotify"
)

var (
	sleeptime int
	path      string
	cmd       string
	args      []string
)

func init() {
	flag.IntVar(&sleeptime, "t", 30, "-t=30")
	flag.StringVar(&path, "p", "./", "-p=filepath or dirpath")
	flag.StringVar(&cmd, "c", "", "-c=command")
	str := flag.String("a", "", `-a="args1 args2"`)
	flag.Parse()
	args = strings.Split(*str, " ")
}

func main() {
	Watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Init monitor error: ", err.Error())
		return
	}
	if err := Watch.Add(path); err != nil {
		log.Println("Add monitor path error: ", path)
		return
	}
	var (
		cron bool = false
		lock      = new(sync.Mutex)
	)
	for {
		select {
		case event := <-Watch.Events:
			log.Printf("Monitor event %s", event.String())
			if !cron {
				cron = true
				go func() {
					T := time.After(time.Second * time.Duration(sleeptime))
					<-T
					if err := call(cmd, args...); err != nil {
						log.Println(err)
					}
					lock.Lock()
					cron = false
					lock.Unlock()
				}()
			}
		case err := <-Watch.Errors:
			log.Println(err)
			return
		}
	}
}

func call(programe string, args ...string) error {
	cmd := exec.Command(programe, args...)
	buf, err := cmd.Output()
	if err != nil {
		return err
	}
	log.Printf("\n%s\n", string(buf))
	return nil
}