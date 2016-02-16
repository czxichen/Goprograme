package log

import (
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

type Log struct {
	logs  *log.Logger
	level int
	io.Closer
}

func NewLog(logpath string, level int) *Log {
	file, err := os.OpenFile(logpath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println("Error", err)
		log.Println("Error", "日志输出到标准输出.")
	}
	var l *log.Logger = log.New(os.Stdout, now(), 0)
	if file != nil {
		l = log.New(file, now(), 0)
		go flushLogFile(file)
	}
	file.Seek(0, 2)
	return &Log{l, level, file}
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05 ")
}

func flushLogFile(File *os.File) {
	for _ = range time.NewTicker(300 * time.Second).C {
		if File == nil {
			return
		}
		File.Sync()
	}
}

func (self *Log) SetLogLevel(level int) {
	if level > 4 {
		return
	}
	self.level = level
}

func (self *Log) Printf(formate string, v ...interface{}) {
	self.logs.Printf(formate, v)
}

func (self *Log) PrintfI(formate string, v ...interface{}) {
	if self.level > 1 {
		return
	}
	self.logs.Printf("Info-> "+formate, v...)
}

func (self *Log) PrintfW(formate string, v ...interface{}) {
	if self.level > 2 {
		return
	}
	self.logs.Printf("Warn-> "+formate, v...)
}

func (self *Log) PrintfE(formate string, v ...interface{}) {
	if self.level > 3 {
		return
	}
	self.logs.Printf("Error-> "+formate, v...)
}

func (self *Log) PrintfF(formate string, v ...interface{}) {
	if self.level > 4 {
		return
	}
	self.logs.Fatalf("Fatal-> "+formate, v...)
}

func (self *Log) InfoPrintf(callers int, formate string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(callers + 1)
	if !ok {
		return
	}
	self.logs.Printf("File->%s Line->%d\n", file, line)
	self.logs.Printf(formate, v...)
}
