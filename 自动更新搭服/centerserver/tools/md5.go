package tools

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func Md5(path string) (string, error) {
	File, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer File.Close()
	m := md5.New()
	io.Copy(m, File)
	return fmt.Sprintf("%X", string(m.Sum([]byte{}))), nil
}
