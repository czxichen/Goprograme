package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func main() {
	lis, err := net.Listen("tcp", ":1789")
	if err != nil {
		return
	}
	defer lis.Close()

	srv := rpc.NewServer()
	if err := srv.RegisterName("Json", new(Json)); err != nil {
		return
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("lis.Accept(): %v\n", err)
			continue
		}
		go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

type Json struct {
	Name string `json:name`
	Age  int    `json:age`
}

func (self *Json) Getname(args Json, result *Json) error {
	if args.Name == "di" {
		log.Println(args)
		*result = Json{"Xichen", 24}
		return nil
	}
	return errors.New("Input Name")
}
