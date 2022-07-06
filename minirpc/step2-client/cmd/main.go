package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"minirpc"
	"minirpc/codec"
)

func startServer(addr chan string) {
	defer func() {
		log.Println("rpc server closed")
	}()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalln("listen err: ", err)
	}
	log.Println("rpc server is starting on ", l.Addr())
	addr <- l.Addr().String()
	minirpc.Accept(l)
}

func main() {
	log.SetFlags(3)
	addr := make(chan string)
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	_ = json.NewEncoder(conn).Encode(minirpc.DefaultOption)
	time.Sleep(time.Second * 3)
	cc := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{ServiceMethod: "Foo.Sum", Seq: uint64(i)}
		_ = cc.Write(h, fmt.Sprintf("minirpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply: ", reply)
	}
}
