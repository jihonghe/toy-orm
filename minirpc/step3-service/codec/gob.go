package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // close the connection by server or client
	buf  *bufio.Writer      // a buffered writer for performance
	dec  *gob.Decoder       // decoder for request or response
	enc  *gob.Encoder       // encoder for request or response
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(header *Header) (err error) {
	return g.dec.Decode(header)
}

func (g *GobCodec) ReadBody(body interface{}) (err error) {
	return g.dec.Decode(body)
}

func (g *GobCodec) Write(header *Header, content interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()

	if err = g.enc.Encode(header); err != nil {
		log.Println("rpc codec: gob err encoding header, err: ", err)
		return
	}
	if err = g.enc.Encode(content); err != nil {
		log.Println("rpc codec: gob err encoding body, err: ", err)
		return
	}

	return
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) (c Codec) {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
