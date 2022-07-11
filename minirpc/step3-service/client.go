package minirpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"minirpc/codec"
)

var (
	ErrorShutdown  = fmt.Errorf("connection is shutdown")
	ErrorNilOption = fmt.Errorf("nil Option error")
)

type Call struct {
	Seq           uint64      // the sequence number of the request
	ServiceMethod string      // the service of the request calls
	Args          interface{} // the request body for service
	Reply         interface{} // the response body for service
	Error         error       // the error that replied by service or client
	Done          chan *Call  // strobes when call is completed
}

func (c *Call) done() {
	c.Done <- c
}

type Client struct {
	opt      *Option          // the codec method that will be sent to server when connected to server
	cc       codec.Codec      // used for codec the request and response
	mu       sync.Mutex       // lock when call some methods
	sending  sync.Mutex       // lock before a request is completed
	header   codec.Header     // reused when sending request
	seq      uint64           // the sequence number for next registered request
	pending  map[uint64]*Call // the waiting list of request
	closing  bool             // closed by user
	shutdown bool             // closed when server notify
}

func NewClient(conn net.Conn, opt *Option) (client *Client, err error) {
	if opt == nil {
		return nil, ErrorNilOption
	}
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		return nil, fmt.Errorf("unsupported codec type '%s'", opt.CodecType)
	}
	err = json.NewEncoder(conn).Encode(opt)
	if err != nil {
		return nil, fmt.Errorf("rpc client: send option %v failed, err: %s", *opt, err.Error())
	}
	client = &Client{
		seq: 1,
		opt: opt,
		cc:  codecFunc(conn),
	}
	go client.receive()
	return
}

func Dial(network, addr string, opt *Option) (client *Client, err error) {
	if opt == nil || opt.CodecType == "" {
		opt = &DefaultOption
	}
	opt.MagicNumber = MagicNumber
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	// notes: where to create, where to release
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

var _ io.Closer = (*Client)(nil)

func (c *Client) Close() (err error) {
	c.lock()
	defer c.unlock()
	if c.shutdown {
		return ErrorShutdown
	}
	c.closing = true
	return c.cc.Close()
}

func (c *Client) IsAvailable() (available bool) {
	c.lock()
	defer c.unlock()
	return !c.closing && !c.shutdown
}

func (c *Client) Go(ServiceMethod string, args, reply interface{}, done chan *Call) (call *Call) {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call = &Call{
		ServiceMethod: ServiceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)

	return call
}

func (c *Client) Call(ServiceMethod string, args, reply interface{}) (err error) {
	call := <-c.Go(ServiceMethod, args, reply, nil).Done
	return call.Error
}

func (c *Client) send(call *Call) {
	c.sendLock()
	defer c.sendUnlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""
	err = c.cc.Write(&c.header, call.Args)
	if err != nil {
		removeCall := c.removeCall(call.Seq)
		if removeCall != nil {
			removeCall.Error = err
			removeCall.done()
		}
	}
}

func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			err = c.cc.ReadBody(nil)
			call.Error = fmt.Errorf(h.Error)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = fmt.Errorf("reading body failed, err: %s", err.Error())
			}
			call.done()
		}
	}
	c.terminateCalls(err)
}

func (c *Client) registerCall(call *Call) (callSeq uint64, err error) {
	c.lock()
	defer c.unlock()
	if c.closing || c.shutdown {
		return 0, ErrorShutdown
	}

	call.Seq = c.seq
	callSeq = call.Seq
	if c.pending == nil {
		c.pending = make(map[uint64]*Call)
	}
	c.pending[call.Seq] = call
	c.seq++

	return
}

func (c *Client) removeCall(seq uint64) (call *Call) {
	c.lock()
	defer c.unlock()
	call = c.pending[seq]
	delete(c.pending, seq)
	return
}

// terminateCalls will remove the call from Client.pending and call Client.done() to notify the call with err
func (c *Client) terminateCalls(err error) {
	c.sendLock()
	defer c.sendUnlock()
	c.lock()
	defer c.unlock()
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

func (c *Client) lock() {
	c.mu.Lock()
}

func (c *Client) unlock() {
	c.mu.Unlock()
}

func (c *Client) sendLock() {
	c.sending.Lock()
}

func (c *Client) sendUnlock() {
	c.sending.Unlock()
}
