package codec

import "io"

type Header struct {
	ServiceMethod string //服务名
	Seq           uint64 //请求ID
	Error         string
}

//编解码接口

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// Codec 的构造函数，客户端和服务端可以通过 Codec 的 Type 得到构造函数
type NewCodecFunc func(io.ReadWriteCloser) Codec

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
