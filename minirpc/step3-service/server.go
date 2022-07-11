package minirpc

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

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

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Register(recv interface{}) (err error) {
	svc := newService(recv)
	_, dup := s.serviceMap.LoadOrStore(svc.name, svc)
	if dup {
		return fmt.Errorf("rpc server: service %s is already defined", svc.name)
	}
	return
}

func Register(recv interface{}) (err error) {
	return DefaultServer.Register(recv)
}

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
	header      *codec.Header
	argv, reply reflect.Value
	mType       *methodType
	svc         *service
}

func (r *request) String() string {
	return fmt.Sprintf("{header: %+v, argv: %+v, reply: %+v}", *r.header, r.argv, r.reply)
}

func (s *Server) findService(serviceMethod string) (svc *service, mType *methodType, err error) {
	dotIdx := strings.LastIndex(serviceMethod, ".")
	if dotIdx < 0 {
		return nil, nil, fmt.Errorf("rpc server: illegle format service/method %s", serviceMethod)
	}
	svcName, methodName := serviceMethod[:dotIdx], serviceMethod[dotIdx+1:]
	svcInter, ok := s.serviceMap.Load(svcName)
	if !ok {
		return nil, nil, fmt.Errorf("rpc server: service %s not found", svcName)
	}
	svc = svcInter.(*service)
	mType = svc.method[methodName]
	if mType == nil {
		return nil, nil, fmt.Errorf("rpc server: method %s of service %s not found", methodName, svcName)
	}
	return
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	header := codec.Header{}
	err := cc.ReadHeader(&header)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header failed, err: ", err)
		}
		return nil, err
	}
	return &header, nil
}

func (s *Server) readRequest(cc codec.Codec) (req *request, err error) {
	header, err := s.readRequestHeader(cc)
	if err != nil {
		return
	}
	req = &request{header: header}
	req.svc, req.mType, err = s.findService(req.header.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mType.newArgv()
	req.reply = req.mType.newReplyV()
	argvI := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvI = req.argv.Addr().Interface()
	}
	err = cc.ReadBody(argvI)
	if err != nil {
		log.Println("rpc server: read argv failed, err: ", err)
		return req, err
	}
	return req, nil
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	err := req.svc.call(req.mType, req.argv, req.reply)
	if err != nil {
		req.header.Error = err.Error()
		s.sendResponse(cc, req.header, invalidRequest, sending)
		return
	}
	s.sendResponse(cc, req.header, req.reply.Interface(), sending)
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	err := cc.Write(h, body)
	if err != nil {
		log.Println("rpc server: write response failed, err: ", err)
	}
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *methodType) NumCalls() (count uint64) {
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv create the method arg value based on reflect.Type
func (m *methodType) newArgv() (argV reflect.Value) {
	if m.ArgType.Kind() == reflect.Ptr {
		argV = reflect.New(m.ArgType.Elem())
	} else {
		argV = reflect.New(m.ArgType).Elem()
	}
	return
}

func (m *methodType) newReplyV() (replyV reflect.Value) {
	replyV = reflect.New(m.ReplyType.Elem())
	// todo: why reply type is map or slice?
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyV.Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyV.Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return
}

type service struct {
	name     string
	method   map[string]*methodType
	recvType reflect.Type  // the service method`s receiver type
	recvIns  reflect.Value // the service method`s receiver value
}

func newService(recv interface{}) (s *service) {
	s = new(service)
	s.recvType = reflect.TypeOf(recv)
	s.recvIns = reflect.ValueOf(recv)
	s.name = s.recvType.Elem().Name()
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not valid service name", s.name)
	}
	s.registerMethods()
	return
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.recvType.NumMethod(); i++ {
		method := s.recvType.Method(i)
		mType := method.Type
		// Service method has 3 params(service method receiver, method request, method response)
		// and has 1 return(error)
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		// validate if the return value is error
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		// validate the second and third param
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("register method: %s.%s(req %s, resp %s)", s.name, method.Name, argType.String(), replyType.String())
	}
}

// isExportedOrBuiltinType
// builtin type`s package is empty
// exported type`s name is Upper
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argV, replyV reflect.Value) (err error) {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returns := f.Call([]reflect.Value{s.recvIns, argV, replyV})
	errInter := returns[0].Interface()
	if errInter != nil {
		return errInter.(error)
	}
	return nil
}
