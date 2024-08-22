package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// ReadBody implements Codec.
func (g *GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}

// ReadHeader implements Codec.
func (g *GobCodec) ReadHeader(h *Header) error {
	return g.dec.Decode(h)
}

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}

var _ Codec = (*GobCodec)(nil)
