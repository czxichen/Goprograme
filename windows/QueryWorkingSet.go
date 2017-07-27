package main

/*
#include <windows.h>
#include <Psapi.h>

void CountPrivate(PSAPI_WORKING_SET_INFORMATION* workSetInfo,int* count)
{
 	int workSetPrivate = 0;
    for (ULONG_PTR i = 0; i < workSetInfo->NumberOfEntries; ++i)
    {
        if(!workSetInfo->WorkingSetInfo[i].Shared) // 如果不是共享页计数器+1
            workSetPrivate += 1;
    }
	*count = workSetPrivate;
}
*/
import "C"

import (
	"flag"
	"fmt"
	"syscall"
	"unsafe"
)

const LARGE_BUFFER_SIZE = 256 * 1024 * 1024

var (
	workSetInfo     C.PSAPI_WORKING_SET_INFORMATION
	psapi           = syscall.NewLazyDLL("Psapi.dll")
	QueryWorkingSet = psapi.NewProc("QueryWorkingSet")
)

func main() {
	pid := flag.Int("p", 0, "-p pid")
	flag.Parse()

	h, err := syscall.OpenProcess(0X0400|0X0010, false, uint32(*pid))
	if err != nil {
		return
	}
	defer syscall.CloseHandle(h)
	var size = Query(h) * syscall.Getpagesize() / 1024

	//There is no panic error using 'println("Memory:", size)' here
	fmt.Println("Memory:", size)
}

func Query(handle syscall.Handle) int {
	var bufferSize = 0x8000
	var buffer = make([]byte, bufferSize)
	for {
		r, _, _ := QueryWorkingSet.Call(uintptr(handle), uintptr(unsafe.Pointer(&buffer[0])), uintptr(bufferSize))
		if r == 0 {
			bufferSize *= 2
			if bufferSize > LARGE_BUFFER_SIZE {
				return 0
			}
			buffer = make([]byte, bufferSize)
			continue
		}
		break
	}

	var count C.int
	C.CountPrivate((*C.PSAPI_WORKING_SET_INFORMATION)(unsafe.Pointer(&buffer[0])), &count)
	return int(count)
}
