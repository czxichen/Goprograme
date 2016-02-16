package tools

import (
	"archive/zip"
)

func CheckValidZip(path string) bool {
	z, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	z.Close()
	return true
}

func CheckMd5(path string) bool {
	return true
}
