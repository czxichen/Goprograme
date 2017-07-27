package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

/*
#include <windows.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#define SN_ARRAY_COUNT(ary) sizeof(ary)/sizeof(ary[0])
#define MAXCOUNT 128

struct book_data_t
{
	unsigned int nInuse;
	unsigned int nProcId;
	unsigned int nState;
	unsigned int nHash;
	char strName[128];
	char strAddr[128];
	char strPath[256];
	char strInfo[256];
};

struct app_book_t
{
	struct book_data_t BookDatas[MAXCOUNT];
};

struct Fd
{
	HANDLE File;
	HANDLE Map;
}fds;

struct Fd* OpenShare(const char * name)
{
	struct Fd *fd=&fds;
	HANDLE file,mem;
	file = OpenFileMappingA(FILE_MAP_ALL_ACCESS, FALSE, name);
	if (NULL == file)
	{
		printf("1");
		return NULL;
	}
	mem =  MapViewOfFile(file,FILE_MAP_ALL_ACCESS, 0, 0, 0);
	if (NULL == mem)
	{
		printf("2");
		CloseHandle(file);
		fd->File = NULL;
		return NULL;
	}
	fd->File=file;
	fd->Map=mem;

	struct app_book_t* appBook =  (struct app_book_t*)(mem);
	int num = SN_ARRAY_COUNT(appBook->BookDatas);
	printf("max app count=%d\n", num);
	for(int i=0; i<num; i++)
	{
		printf("Name=%s, Info=%s \n",appBook->BookDatas[i].strName,	appBook->BookDatas[i].strInfo);
	}
	return fd;
}

int Close(struct Fd *fd)
{
	if (fd->Map != NULL)
	{
		UnmapViewOfFile(fd->Map);
		fd->Map = NULL;
	}
	if (fd->File != NULL)
	{
		CloseHandle(fd->File);
	}
	return 0;
}
*/
import "C"

type book_data_t struct {
	nInuse  uint32
	nProcId uint32
	nState  uint32
	nHash   uint32
	strName [128]byte
	strAddr [128]byte
	strPath [256]byte
	strInfo [256]byte
}

type app_book_t struct {
	BookDatas [5]book_data_t
}

var (
	kernel32   = syscall.NewLazyDLL("Kernel32.dll")
	OpenMutexA = kernel32.NewProc("OpenMutexA")
)

const (
	MUTEX_ALL_ACCESS = 0x1F0001
)

func main() {
	fmt.Println("This")
	cname := C.CString(os.Args[1])
	if h := C.OpenShare(cname); h != nil {
		b := *(*app_book_t)(unsafe.Pointer(h.Map))
		list := b.BookDatas[:]
		for _, v := range list {
			fmt.Println("nInuse:", v.nInuse)
			fmt.Println("nHash:", v.nHash)
			fmt.Println("nProcId:", v.nProcId)
			fmt.Println("nState:", v.nState)
			fmt.Println("strName:", string(v.strName[:]))
			//			fmt.Println("strAddr:", syscall.UTF16ToString(v.strAddr[:]))
			//			fmt.Println("strPath:", syscall.UTF16ToString(v.strPath[:]))
			//			fmt.Println("strInfo:", syscall.UTF16ToString(v.strInfo[:]))
		}
		C.Close(h)
	}
	fmt.Println("Here")
	C.free(unsafe.Pointer(cname))
}

func Exist(name string) bool {
	cname := unsafe.Pointer(C.CString("mutex_" + name))
	defer C.free(cname)
	r, _, _ := OpenMutexA.Call(MUTEX_ALL_ACCESS, 0, uintptr(cname))
	if r == 0 {
		return false
	} else {
		syscall.CloseHandle(syscall.Handle(r))
		return true
	}
}
