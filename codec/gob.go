package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser //conn 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例
	buf  *bufio.Writer      //buf 是为了防止阻塞而创建的带缓冲的 Writer
	dec  *gob.Decoder       //解码
	enc  *gob.Encoder       //编码
}

// Close implements Codec
func (c *GobCodec) Close() error {
	return c.conn.Close()
}

// ReadBody implements Codec
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

// ReadHeader implements Codec
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

// Write implements Codec
func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()

	if err := c.enc.Encode(h); err != nil {
		log.Println("rpg codec: gob error encodeing header: ", err)
		return err
	}

	if err := c.enc.Encode(body); err != nil {
		log.Println("rpg codec: gob error encoding body: ", err)
		return err
	}

	return nil
}

// 这是确保接口被实现常用的方式。即利用强制类型转换，
// 确保 struct GobCodec 实现了接口 Codec。
// 这样 IDE 和编译期间就可以检查，而不是等到使用的时候。
var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)

	//返回初始化
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
