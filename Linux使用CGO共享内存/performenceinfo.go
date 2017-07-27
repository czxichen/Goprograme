package main

import (
	"os"
	"unsafe"
)

/*
#include <stdio.h>
#include <windows.h>
#include <Winuser.h>
#include <stdlib.h>
#include <string.h>


typedef struct EnumFunArg
{
    HWND      hWND;
    DWORD    dwProcessId;
}EnumFunArg,*LPEnumFunArg;

BOOL CALLBACK lpEnumFunc(HWND hwnd, LPARAM lParam)
{
    EnumFunArg  *pArg = (LPEnumFunArg)lParam;
    DWORD  processId;
    GetWindowThreadProcessId(hwnd, &processId);
    if( processId == pArg->dwProcessId)
    {
        pArg->hWND = hwnd;
        return FALSE;
    }
    return TRUE;
}

int ReturnWnd(DWORD processID)
{
   BOOL re = FALSE;
   EnumFunArg wi;
   wi.dwProcessId = processID;
   wi.hWND   =  NULL;
   EnumWindows(lpEnumFunc,(LPARAM)&wi);
   if(wi.hWND)
   {
		if (IsHungAppWindow(wi.hWND))
		{
			return 1;
		}
   }
   else
   {
		return -1;
   }
	return 0;
}

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
		return NULL;
	}
	mem =  MapViewOfFile(file,FILE_MAP_ALL_ACCESS, 0, 0, 0);
	if (NULL == mem)
	{
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

func main() {
	name := C.CString(os.Args[1])
	fd := C.OpenShare(name)
	C.Close(fd)
	C.free(unsafe.Pointer(name))
}
