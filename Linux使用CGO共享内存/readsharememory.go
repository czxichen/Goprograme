package main

import (
	"flag"
	"fmt"
)

/*
#include <stdio.h>
#include <sys/shm.h>
#include <sys/ipc.h>
char* GetMem(int Shm_id)
{
    char *Shm_add;
    Shm_add=shmat(Shm_id,0,0);
    if (Shm_add==(void *)-1)
    {
        return (char *)-1;
    }
    return (char *)Shm_add;
}
int close(char *Shm_add)
{
	int result;
	result=shmdt(Shm_add);
    if (result != 0)
    {
        return result;
    }
}
*/
import "C"

type ShareMem struct {
	Addr *C.char
}

func main() {
	i := flag.Int("m", 0, "-m=12345")
	flag.Parse()
	x := GetShareMem(*i)
	defer x.Close()
	fmt.Println(x.Read())
}

func GetShareMem(shm_id int) *ShareMem {
	m := C.GetMem(C.int(shm_id))
	return &ShareMem{m}
}

func (self *ShareMem) Close() {
	C.close(self.Addr)
}

func (self *ShareMem) Read() string {
	return C.GoString(self.Addr)
}
