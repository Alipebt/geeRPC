# go语言实现RPC框架

## 1. 服务端与消息编码

### 1.1 消息的序列化与反序列化

`Header struct`是经典`RPC`调用所需要的参数，分为`服务名和方法名`，`请求ID`，`错误信息`三个字段，其中`请求ID`用来区分不同的请求，之后需要编解码的数据即为`Header`和另一个为`infterface`的主体

`Codec`接口是处理`Header struct`的函数集合，其构造函数`NewCodecFunc`需传入`io.ReadWriteCloser`,由于`Codec`可能处理多个类型的编解码，所以将其存入`NewCodecFuncMap`，即可利用`key`来寻找合适的编解码函数，其中`NewCodecFuncMap[GobType] = NewGobCodec`就是将`GobType`类型的编解码交给`NewGobCodec`处理

```go
// 这是确保接口被实现常用的方式。即利用强制类型转换，
// 确保 struct GobCodec 实现了接口 Codec。
// 这样 IDE 和编译期间就可以检查，而不是等到使用的时候。
var _ Codec = (*GobCodec)(nil)
```



