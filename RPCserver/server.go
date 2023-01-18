package RPCserver

import "RPC/codec"

type Option struct {
	length uint64
	header codec.Type
}
