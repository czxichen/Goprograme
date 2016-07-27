package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func main() {
	section := uintptr(0x0004)
	var InheritHandle = bool
	var inherit = (uintptr)(unsafe.Pointer(&InheritHandle))
	var ShareMemName string = "Watch"
	var sharememname = (uintptr)(unsafe.Pointer(&ShareMemName))
	kernet32 := syscall.NewLazyDLL("kernel32.dll")
	defer syscall.FreeLibrary(kernet32)

	OpenFileMapping := kernet32.NewProc("OpenFileMappingA")
	ptr, _, err := OpenFileMapping.Call(section, inherit, sharememname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer syscall.CloseHandle(syscall.Handle(ptr))
}
