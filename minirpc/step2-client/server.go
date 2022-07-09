package minirpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"

	"minirpc/codec"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int        // it marks this is a minirpc request
	CodecType   codec.Type // the encoding/decoding method that client choose
}

var DefaultOption = Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()
	// select the CodecFunc for the connection base on Option
	var opt Option
	err := json.NewDecoder(conn).Decode(&opt)
	if err != nil {
		log.Println("rpc server: decode Option failed, err: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x, expected %x for rpc", opt.MagicNumber, MagicNumber)
		return
	}
	codeCFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codeCFunc == nil {
		log.Printf("rpc server: unsupported codec type %s, only support gob", opt.CodecType)
		return
	}
	log.Printf("recv a new req for a codec: %v", opt)
	s.serveCodec(codeCFunc(conn))
}

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept failed, err: ", err)
			return
		}
		log.Println("recv a new connection")
		go s.ServeConn(conn)
	}
}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc)
		log.Print("recv a new req")
		if err != nil {
			if req == nil {
				break
			}
			req.header.Error = err.Error()
			s.sendResponse(cc, req.header, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	header       *codec.Header
	argv, replyv reflect.Value
}

func (s *Server) readRequestHeader(cc codec.Codec) (header *codec.Header, err error) {
	header = &codec.Header{}
	err = cc.ReadHeader(header)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header failed, err: ", err)
		}
		return
	}
	return
}

func (s *Server) readRequest(cc codec.Codec) (req *request, err error) {
	header, err := s.readRequestHeader(cc)
	if err != nil {
		return
	}
	req = &request{header: header}
	req.argv = reflect.New(reflect.TypeOf(""))
	err = cc.ReadBody(req.argv.Interface())
	if err != nil {
		log.Println("rpc server: read argv failed, err: ", err)
	}
	return
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	err := cc.Write(h, body)
	if err != nil {
		log.Println("rpc server: write response failed, err: ", err)
	}
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("req-header: ", req.header, " req-body: ", req.argv)
	req.replyv = reflect.ValueOf(fmt.Sprintf("minirpc resp %d", req.header.Seq))
	s.sendResponse(cc, req.header, req.replyv.Interface(), sending)
}
