# go语言实现RPC框架

## 1. 服务端与消息编码

### 1.1 消息的序列化与反序列化

`Header struct`是经典`RPC`调用所需要的参数，分为`服务名和方法名`，`请求ID`，`错误信息`三个字段，其中`请求ID`用来区分不同的请求，之后需要编解码的数据即为`Header`和另一个为`infterface`的主体

`Codec`接口是处理`Header struct`的函数集合，其构造函数`NewCodecFunc`需传入`io.ReadWriteCloser`,由于`Codec`可能处理多个类型的编解码，所以将其存入`NewCodecFuncMap`，便可利用`key`来寻找合适的编解码函数，其中`NewCodecFuncMap[GobType] = NewGobCodec`就是将`GobType`类型的编解码交给`NewGobCodec`处理初始化

```go
// 这是确保接口被实现常用的方式。即利用强制类型转换
// 确保 struct GobCodec 实现了接口 Codec
// 这样 IDE 和编译期间就可以检查，而不是等到使用的时候
var _ Codec = (*GobCodec)(nil)
```

### 1.2 通信过程

客户端与服务端通信一般需要协商一些内容，例如 HTTP 报文，分为 `header` 和 `body` 两部分，`body` 的格式和长度通过 `header` 中的 `Content-Type` 和 `Content-Length` 指定，服务端通过解析 `header` 就能够知道如何从 `body` 中读取需要的信息。对于 `RPC` 协议来说，这部分协商是需要自主设计的。一般在报文的最开始会规划固定的字节，来协商相关的信息。比如第1个字节用来表示序列化方式，第2个字节表示压缩方式，第3-6字节表示 header 的长度，7-10 字节表示 body 的长度。

```
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------> | <-------   编码方式由 CodeType 决定  ------->|
```

在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文可能是这样的：

```
| Option | Header1 | Body1 | Header2 | Body2 | ...
```

### 1.3 服务端的实现

首先需要一个`server`结构体，不需要任何成员字段，通过`NewServer()`创建一个默认`server`实例。

`Accept()`是`server`的一个方法，需要`net.Listener`作为参数。内部通过`for`循环等待`socket`建立连接，并开启协程处理，处理过程交给了`ServerConn()`方法

启动方法很简单：

```go
lis,_ := net.Listen("tcp",":9999")
geerpc.Accept(lis)
```

`ServeConn` 的实现就和之前讨论的通信过程紧密相关了，首先使用 `json.NewDecoder` 反序列化得到 Option 实例，检查 `MagicNumber` 和` CodeType `的值是否正确。然后根据`CodeType`得到对应的消息编解码器，接下来的处理交给 `serverCodec`。

`serveCodec` 的过程非常简单。主要包含三个阶段

- 读取请求 `readRequest`
- 处理请求 `handleRequest`
- 回复请求 `sendResponse`

在一次连接中，允许接收多个请求，即多个` request header `和 `request body`，因此这里使用了 `for` 无限制地等待请求的到来，直到发生错误（例如连接被关闭，接收到的报文有问题等），这里需要注意的点有三个：

- `handleRequest` 使用了协程并发执行请求。
- 处理请求是并发的，但是回复请求的报文必须是逐个发送的，并发容易导致多个回复报文交织在一起，客户端无法解析。在这里使用锁(`sending`)保证。

`New`接口：

```go
func New(typ Type ) value
```

`New` 返回一个 `Value`，表示指向指定类型的新零值的指针。也就是说，返回值的类型是 `PointerTo(typ)`。

`func (Value) Inetrface` ：

接口将`v` 的当前值作为接口{}返回。它相当于：

```go
var i interface{} = (v 的基础值)
```

如果 `Value` 是通过访问未导出的结构字段获得的，它会发生恐慌。

`func (Value) Elem` ：

```go
func (v Value) Elem() Value
```

Elem 返回接口 v 包含的值或指针 v 指向的值。如果 v 的种类不是接口或指针，它会发生恐慌。如果 v 为 nil，则返回零值。

`func Dial`：

```go
func Dial(network, address string) (Conn, error)
```

拨号(`Dial`)连接到指定网络上的地址。
