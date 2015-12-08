package tool

import (
	"io/ioutil"
)

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
