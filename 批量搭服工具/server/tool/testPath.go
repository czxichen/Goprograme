package tool

import (
	"fmt"
	"os"
)

func TestPath(path string) {
	info, err := os.Lstat(path)
	if err != nil {
		fmt.Printf("check %s .error_info :%s", path, err)
		os.Exit(9)
	}
	if info.IsDir() {
		fmt.Printf("check %s is directory", path)
		os.Exit(10)
	}
	return
}
