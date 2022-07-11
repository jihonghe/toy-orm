package codec

import (
	"io"
)

type Header struct {
	ServiceMethod string // format "Service.method"
	Seq           uint64 // sequence number of a request
	Error         string // if error occurs in server, then set the error
}

type Codec interface {
	io.Closer
	ReadHeader(header *Header) (err error)
	ReadBody(body interface{}) (err error)
	Write(header *Header, content interface{}) (err error)
}

type NewCodecFunc func(closer io.ReadWriteCloser) (c Codec)

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
