package slog

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Logger struct {
	console bool
	warn    bool
	info    bool
	tformat func() string
	file    chan string
}

func ce() {
	fmt.Println("FUCK")
}
func NewLog(level string, console bool, File *os.File, buf int) (*Logger, error) {
	log := &Logger{console: console, tformat: format}
	if File != nil {
		FileInfo, err := File.Stat()
		if err != nil {
			return nil, err
		}
		mode := strings.Split(FileInfo.Mode().String(), "-")
		if strings.Contains(mode[1], "w") {
			str_chan := make(chan string, buf)
			log.file = str_chan
			go func() {
				for {
					fmt.Fprintln(File, <-str_chan)
				}
			}()
			defer func() {
				for len(str_chan) > 0 {
					time.Sleep(1e9)
				}
			}()
		} else {
			return nil, errors.New("can't write.")
		}
	}
	switch level {
	case "Warn":
		log.warn = true
		return log, nil
	case "Info":
		log.warn = true
		log.info = true
		return log, nil
	}
	return nil, errors.New("level must be Warn or Info.")
}

func (self *Logger) Error(info interface{}) {
	if self.console {
		fmt.Println("Error", self.tformat(), info)
	}
	if self.file != nil {
		self.file <- fmt.Sprintf("Error %s %s", self.tformat(), info)

	}
}

func (self *Logger) Warn(info interface{}) {
	if self.warn && self.console {
		fmt.Println("Warn", self.tformat(), info)
	}
	if self.file != nil {
		self.file <- fmt.Sprintf("Warn %s %s", self.tformat(), info)
	}
}
func (self *Logger) Info(info interface{}) {
	if self.info && self.console {
		fmt.Println("Info", self.tformat(), info)
	}
	if self.file != nil {
		self.file <- fmt.Sprintf("Info %s %s", self.tformat(), info)
	}
}
func (self *Logger) Close() {
	for len(self.file) > 0 {
		time.Sleep(1e8)
	}
}
func format() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
