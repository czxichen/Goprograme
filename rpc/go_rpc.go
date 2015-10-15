package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
)

func s_main() {
	rpc.Register(new(Remote))
	rpc.HandleHTTP()
	http.ListenAndServe(":1789", nil)
}
func main() {
	args := Test{"di", 24, []string{"basketball", "buitifulgirl"}}
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1789")
	if err != nil {
		fmt.Println(err)
		return
	}
	var b bool
	err = client.Call("Remote.GetInfo", args, &b)
	if err != nil {
		fmt.Println(err)
	}
	var s string
	err = client.Call("Remote.GetName", "WaCao", &s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(b, s)
}

type Remote int32

type Test struct {
	Name  string
	Age   int
	Hobby []string
}

func (x *Remote) GetName(args string, result *string) error {
	if args != "" {
		*result = args
		return nil
	}
	return errors.New("Input Empty")
}
func (x *Remote) GetInfo(args Test, result *bool) error {
	if len(args.Hobby) == 0 {
		return errors.New("Hobby Is Empty")
	}
	fmt.Println(args)
	*result = true
	return nil
}