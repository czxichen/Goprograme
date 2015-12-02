package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

/*
#cgo linux LDFLAGS: -lrt
#include <stdio.h>
#include <sys/shm.h>
#include <stdlib.h>
#include <sys/ipc.h>
#include <string.h>

int Size;
int Shm_id;
key_t Key;
char *Shm_add;
int getMem(char *pathname,int size)
{
    Key = ftok((char *)pathname,0x03);
    if(Key==-1)
    {
        return -1;
    }
    Shm_id=shmget(Key,size,IPC_CREAT|IPC_EXCL|0600);
    if (Shm_id==-1)
    {
        return -2;
    }
    Shm_add=shmat(Shm_id,0,0);
    if (Shm_add==(void *)-1)
    {
	    return -3;
    }
    Size=size;
    return 0;
}
int close()
{
    int result;
    result=shmdt(Shm_add);
    if (result != 0)
    {
        return result;
    }
    result=shmctl(Shm_id, IPC_RMID, NULL);
    if (result != 0)
    {
        return result;
    }
    return 0;
}
int write(char *str)
{
        int len;
	len=strlen(str);
	if (len >= Size)
	{
		return 0;
	}
	strcpy((char *)Shm_add,(char *)str);
	return len;
}
char* read()
{
	return (char *)Shm_add;
}
*/
import "C"

type Mem struct {
	Shm_id C.int
	size   int
}

func main() {
	m, err := GetShareMem("/tmp/.tmpfile", 1024)
	if err != nil {
		fmt.Println(err)
	}
	defer m.Close()
	m.Write("czxichen")
	fmt.Println(m.Read())
	time.Sleep(60e9)
}

func GetShareMem(path string, size int) (*Mem, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("Must Is File")
	}
	num := C.getMem(C.CString(path), C.int(size))

	switch num {
	case 0:
		return &Mem{C.Shm_id, size}, nil
	case -1:
		return nil, errors.New("Get Key error")
	case -2:
		return nil, errors.New("Get Shm_id error")
	case -3:
		C.close()
		return nil, errors.New("Map AddrSpace error")
	}
	C.close()
	return nil, errors.New("Unknow error")
}

func (self *Mem) Close() error {
	num := C.close()
	if num != 0 {
		cmd := exec.Command("ipcrm", "-m", fmt.Sprint(self.Shm_id))
		return cmd.Run()
	}
	return nil
}
func (self *Mem) Read() string {
	return C.GoString(C.read())
}
func (self *Mem) Write(body string) C.int {
	if len(body) >= self.size {
		return C.int(0)
	}
	return C.write(C.CString(body))
}
