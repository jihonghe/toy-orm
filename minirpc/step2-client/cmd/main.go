package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"minirpc"
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

func connAndSend(reqNum int, network, serverAddr string) {
	client, err := minirpc.Dial(network, serverAddr, nil)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < reqNum; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			args := fmt.Sprintf("minirpc req %d", seq)
			var reply string
			if err = client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatalf("call Foo.Sum failed: %s", err.Error())
			}
			log.Println("recv reply: ", reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(4)
	addr := make(chan string)
	go startServer(addr)
	serverAddr := <-addr

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			connAndSend(5, "tcp", serverAddr)
		}()
	}
	wg.Wait()
}
