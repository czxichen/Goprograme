package tools

import (
	"archive/zip"
	"centerserver/log"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func Unzip(filename, dir string, Log *log.Log) error {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	//fmt.Printf("Info Unzip to %s\n", dir)
	File, err := zip.OpenReader(filename)
	if err != nil {
		//fmt.Printf("Error Open zip faild:\n%s\n", err)
		return errors.New(fmt.Sprintf("Error Open zip faild:\n%s\n", err))
	}
	defer File.Close()
	for _, v := range File.File {
		err := createFile(v, dir)
		if err != nil {
			Log.PrintfE("unzip file err %v \n", err)
			continue
		}
		os.Chtimes(v.Name, v.ModTime(), v.ModTime())
		os.Chmod(v.Name, v.Mode())
		Log.PrintfI("unzip %s %s\n", filename, v.Name)
	}
	return nil
}

func createFile(v *zip.File, dscDir string) error {
	v.Name = dscDir + v.Name
	info := v.FileInfo()
	if info.IsDir() {
		err := os.MkdirAll(v.Name, v.Mode())
		if err != nil {
			return errors.New(fmt.Sprintf("Error Create direcotry %s faild:\n%s\n", v.Name, err))
		}
		return nil
	}
	srcFile, err := v.Open()
	if err != nil {
		return errors.New(fmt.Sprintf("Error Read from zip faild:\n%s\n", err))
	}
	defer srcFile.Close()
	newFile, err := os.Create(v.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("Error Create file faild:\n%s\n", err))
	}
	defer newFile.Close()
	io.Copy(newFile, srcFile)
	return nil
}
