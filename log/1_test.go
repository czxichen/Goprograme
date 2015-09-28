package main

import (
	"fmt"
	"os"
	"slog"
	"testing"
)

func Test_log(T *testing.T) {
	File, _ := os.Create("log")
	log, err := slog.NewLog("Info", false, File, 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer log.Close()
	for i := 0; i < 100000; i++ {
		log.Warn("Nima")
		log.Info("Fuck")
	}

}
func Benchmark_log(b *testing.B) {
	File, _ := os.Create("log")
	log, err := slog.NewLog("Info", false, File, 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer log.Close()
	for i := 0; i < b.N; i++ {
		log.Warn("Nima")
	}
}
