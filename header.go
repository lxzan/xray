package xray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/lxzan/xray/codec"
	"github.com/lxzan/xray/constant"
	"github.com/lxzan/xray/internal"
	"math"
	"net/http"
	"strconv"
)

var (
	binaryLengthEncoding headerLengthEncoding = 2
	textLengthEncoding   headerLengthEncoding = 4

	// TextMapHeader 文本类型头部编码, 4字节, 最大长度=9999
	// text type header code, 4 bytes, max length = 9999
	TextMapHeader = NewHeaderCodec(MapHeaderTemplate, codec.StdJsonCodec).setLengthBytes(textLengthEncoding)

	// BinaryMapHeader 二进制类型头部编码, 2字节, 最大长度=65535
	// binary type header code, 2 bytes, max length = 65535
	BinaryMapHeader = NewHeaderCodec(MapHeaderTemplate, codec.StdJsonCodec).setLengthBytes(binaryLengthEncoding)
)

type (
	Header interface {
		PoolGet() Header                      // 生成一个新的Header, 需要返回指针类型
		PoolPut()                             // 回收资源
		Set(key, value string)                // 设置键值对
		Get(key string) string                // 获取一个值
		Del(key string)                       // 删除
		Len() int                             // 获取长度
		Range(f func(key, value string) bool) // 遍历
	}

	headerLengthEncoding int
)

type MapHeader map[string]string

func (c *MapHeader) PoolGet() Header {
	return HeaderPool().Get(constant.IdMapHeader)
}

func (c *MapHeader) PoolPut() {
	if c.Len() <= 32 {
		for k, _ := range *c {
			delete(*c, k)
		}
		HeaderPool().Put(constant.IdMapHeader, c)
	}
}

func (c *MapHeader) Len() int {
	return len(*c)
}

func (c *MapHeader) Range(f func(key string, value string) bool) {
	for k, v := range *c {
		if !f(k, v) {
			return
		}
	}
}

func (c *MapHeader) Set(key, value string) {
	key = c.formatKey(key)
	(*c)[key] = value
}

func (c *MapHeader) Get(key string) string {
	key = c.formatKey(key)
	return (*c)[key]
}

func (c *MapHeader) Del(key string) {
	key = c.formatKey(key)
	delete(*c, key)
}

func (c *MapHeader) formatKey(key string) string {
	var b = []byte(key)
	var n = len(b)
	for i := 0; i < n-1; i++ {
		if (i == 0 || b[i-1] == '-') && (b[i] >= 'a' && b[i] <= 'z') {
			b[i] -= 32
		}
	}
	return string(b)
}

func NewHttpHeader(h http.Header) *HttpHeader {
	return &HttpHeader{Header: h}
}

type HttpHeader struct {
	http.Header
}

func (c HttpHeader) PoolGet() Header {
	return &HttpHeader{Header: http.Header{}}
}

func (c HttpHeader) PoolPut() {}

func (c HttpHeader) Len() int {
	return len(c.Header)
}

func (c HttpHeader) Range(f func(key string, value string) bool) {
	for k, values := range c.Header {
		for _, v := range values {
			f(k, v)
		}
	}
}

func (c headerLengthEncoding) MaxLength() int {
	if c == binaryLengthEncoding {
		return math.MaxUint16
	}
	return 1e4 - 1
}

type HeaderCodec struct {
	codec          codec.Codec
	lengthEncoding headerLengthEncoding
	template       Header
}

// NewHeaderCodec
// templateId必须是已注册的模板类型
func NewHeaderCodec(template Header, codec codec.Codec) *HeaderCodec {
	return new(HeaderCodec).
		setLengthBytes(textLengthEncoding).
		SetTemplate(template).
		SetCodec(codec)
}

func (c *HeaderCodec) setLengthBytes(lb headerLengthEncoding) *HeaderCodec {
	c.lengthEncoding = lb
	return c
}

// SetCodec 设置头部信息编码
// set the header information code
func (c *HeaderCodec) SetCodec(codec codec.Codec) *HeaderCodec {
	c.codec = codec
	return c
}

// SetTemplate 设置header模板
func (c *HeaderCodec) SetTemplate(template Header) *HeaderCodec {
	c.template = template
	return c
}

func (c *HeaderCodec) Encode(writer *bytes.Buffer, h Header) error {
	var p0 [4]byte
	_, _ = writer.Write(p0[:c.lengthEncoding]) // err must be nil

	if h != nil && h.Len() > 0 {
		if err := c.codec.NewEncoder(writer).Encode(h); err != nil {
			return err
		}
	}

	var headerLength = writer.Len() - int(c.lengthEncoding)
	if headerLength > c.lengthEncoding.MaxLength() {
		return constant.ErrHeaderTooLarge
	}

	var p1 = writer.Bytes()
	var p2 = p1[:c.lengthEncoding]
	if c.lengthEncoding == binaryLengthEncoding {
		binary.BigEndian.PutUint16(p2, uint16(headerLength))
	} else {
		copy(p2, fmt.Sprintf("%04d", headerLength))
	}

	return nil
}

func (c *HeaderCodec) Decode(reader *bytes.Buffer) (Header, error) {
	var v = c.Generate()
	var p [4]byte

	if err := internal.Read(reader, p[:c.lengthEncoding]); err != nil {
		return nil, err
	}

	var headerLength = 0
	if c.lengthEncoding == binaryLengthEncoding {
		headerLength = int(binary.BigEndian.Uint16(p[:c.lengthEncoding]))
	} else {
		n, err := strconv.Atoi(string(p[:c.lengthEncoding]))
		if err != nil {
			return nil, err
		}
		headerLength = n
	}

	if reader.Len() < headerLength {
		return nil, constant.ErrParseHeader
	}

	if headerLength > 0 {
		if err := c.codec.Decode(reader.Next(headerLength), v); err != nil {
			return nil, err
		}
	}
	return v.(Header), nil
}

func (c *HeaderCodec) Generate() Header {
	return c.template.PoolGet()
}
