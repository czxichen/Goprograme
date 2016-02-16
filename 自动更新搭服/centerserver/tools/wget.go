package tools

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
)

type UpdateResult struct {
	Path  string   `json:path`
	Files []string `json:files`
}

func client() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}
func Wget(url, name string) (bool, error) {
	resp, err := client().Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	File, err := os.Create(name)
	if err != nil {
		return false, err
	}
	io.Copy(File, resp.Body)
	File.Close()
	return true, nil
}
