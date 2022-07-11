package main

import (
	"log"
	"net"
	"sync"
	"time"

	"minirpc"
)

type Foo int

type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) (err error) {
	*reply = args.Num1 + args.Num2
	return
}

func startServer(addr chan string) {
	defer func() {
		log.Println("rpc server closed")
	}()
	var foo Foo
	err := minirpc.Register(&foo)
	if err != nil {
		log.Fatalf("register error: %s", err.Error())
	}
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
			args := &Args{Num1: seq, Num2: seq + 1}
			var reply int
			if err = client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatalf("call Foo.Sum failed: %s", err.Error())
			}
			log.Printf("rpc client(%v): %d + %d = %d", &client, args.Num1, args.Num2, reply)
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
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			connAndSend(5, "tcp", serverAddr)
		}()
	}
	wg.Wait()
}
