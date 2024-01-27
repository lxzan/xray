package xray

import (
	"bytes"
	"errors"
	"github.com/lxzan/xray/codec"
	"github.com/lxzan/xray/constant"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func newMapHeader() *MapHeader {
	return MapHeaderTemplate.PoolGet().(*MapHeader)
}

func newContextMocker() *Context {
	var request = &Request{
		Header: HttpHeader{Header: http.Header{}},
		Body:   bytes.NewBuffer(nil),
	}
	var writer = newResponseWriterMocker()
	var ctx = NewContext(request, writer)
	return ctx
}

func newResponseWriterMocker() *responseWriterMocker {
	return &responseWriterMocker{
		protocol:   ProtocolHTTP,
		statusCode: 0,
		header:     HttpHeader{Header: http.Header{}},
		buf:        bytes.NewBuffer(nil),
	}
}

type responseWriterMocker struct {
	protocol   string
	statusCode int
	header     HttpHeader
	buf        *bytes.Buffer
}

func (c *responseWriterMocker) SetProtocol(p string) {
	c.protocol = p
}

func (c *responseWriterMocker) Protocol() string {
	return c.protocol
}

func (c *responseWriterMocker) Header() Header {
	return c.header
}

func (c *responseWriterMocker) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, errors.New("test error")
	}
	return c.buf.Write(p)
}

func (c *responseWriterMocker) Code(code int) {
	c.statusCode = code
}

func (c *responseWriterMocker) Flush() error {
	return nil
}

func (c *responseWriterMocker) Raw() interface{} {
	return nil
}

func TestContext_BindJSON(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var ctx = newContextMocker()
		ctx.Request.Body = bytes.NewBufferString(`{"age":1}`)
		var params = struct {
			Age int `json:"age"`
		}{}
		as.NoError(ctx.BindJSON(&params))
		as.Equal(1, params.Age)
	})

	t.Run("", func(t *testing.T) {
		var ctx = newContextMocker()
		ctx.Request.Body = strings.NewReader(`{"age":1}`)
		var params = struct {
			Age int `json:"age"`
		}{}
		as.NoError(ctx.BindJSON(&params))
		as.Equal(1, params.Age)
	})

	t.Run("", func(t *testing.T) {
		var ctx = newContextMocker()
		ctx.Request.Body = bytes.NewBufferString(`{"age":"1}`)
		var params = struct {
			Age int `json:"age"`
		}{}
		as.Error(ctx.BindJSON(&params))
	})

	t.Run("", func(t *testing.T) {
		var ctx = newContextMocker()
		var params = struct {
			Age int `json:"age"`
		}{}
		as.Error(ctx.BindJSON(&params))
	})
}

func TestContext_Write(t *testing.T) {
	var as = assert.New(t)

	t.Run("write json 1", func(t *testing.T) {
		var ctx = newContextMocker()
		var params = Any{"name": "aha"}
		if err := ctx.WriteJSON(http.StatusOK, params); err != nil {
			as.NoError(err)
			return
		}
		var writer = ctx.Writer.(*responseWriterMocker)
		as.Equal(http.StatusOK, writer.statusCode)
		as.Equal(constant.MimeJson, writer.header.Get(constant.ContentType))
		var buf = bytes.NewBufferString("")
		defaultJsonCodec.NewEncoder(buf).Encode(params)
		as.Equal(buf.Len(), writer.buf.Len())
	})

	t.Run("write json 2", func(t *testing.T) {
		var ctx = newContextMocker()
		var header = &headerMocker{newMapHeader()}
		header.Set(constant.ContentType, constant.MimeJson)
		as.Error(ctx.WriteJSON(http.StatusOK, header))
	})

	t.Run("write string", func(t *testing.T) {
		var ctx = newContextMocker()
		var params = "hello"
		if err := ctx.WriteString(http.StatusOK, params); err != nil {
			as.NoError(err)
			return
		}
		var writer = ctx.Writer.(*responseWriterMocker)
		as.Equal(http.StatusOK, writer.statusCode)
		as.Equal("", writer.header.Get(constant.ContentType))
		as.Equal(params, writer.buf.String())
	})

	t.Run("write string", func(t *testing.T) {
		var ctx = newContextMocker()
		var params = []byte("hello")
		if err := ctx.WriteBytes(http.StatusOK, params); err != nil {
			as.NoError(err)
			return
		}
		var writer = ctx.Writer.(*responseWriterMocker)
		as.Equal(http.StatusOK, writer.statusCode)
		as.Equal("", writer.header.Get(constant.ContentType))
		as.Equal(string(params), writer.buf.String())
	})

	t.Run("write reader 1", func(t *testing.T) {
		var ctx = newContextMocker()
		var header = &headerMocker{newMapHeader()}
		header.Set(constant.ContentType, constant.MimeJson)
		as.Error(ctx.WriteReader(http.StatusOK, header))
	})

	t.Run("write reader 2", func(t *testing.T) {
		var ctx = newContextMocker()
		as.Error(ctx.WriteReader(http.StatusOK, bytes.NewBufferString("")))
	})

	t.Run("write bytes", func(t *testing.T) {
		var ctx = newContextMocker()
		as.Error(ctx.WriteBytes(http.StatusOK, []byte{}))
	})
}

func TestContext_Storage(t *testing.T) {
	var as = assert.New(t)
	var ctx = newContextMocker()
	ctx.Set("name", "aha")
	ctx.Set("age", 1)
	{
		v, _ := ctx.Get("name")
		as.Equal("aha", v)
	}
	{
		v, _ := ctx.Get("age")
		as.Equal(1, v)
	}
}

func TestContext_Others(t *testing.T) {
	var as = assert.New(t)
	var ctx = newContextMocker()
	SetJsonCodec(codec.StdJsonCodec)
	SetBufferPool(newBufferPool())
	as.Nil(ctx.Request.Raw)
	as.Nil(ctx.Writer.Raw())
}

func TestContext_Param(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var ctx = NewContext(&Request{
			Header: NewHttpHeader(http.Header{}),
			VPath:  "/:id",
		}, newResponseWriterMocker())
		id := ctx.Param("id")
		as.Equal("", id)
	})

	t.Run("", func(t *testing.T) {
		var ctx = NewContext(&Request{
			Header: NewHttpHeader(http.Header{}),
			VPath:  "/api/v1",
		}, newResponseWriterMocker())
		id := ctx.Param("id")
		as.Equal("", id)
	})

	t.Run("", func(t *testing.T) {
		var ctx = NewContext(&Request{
			VPath: "/api/v1",
			RPath: "/api/v1",
		}, newResponseWriterMocker())
		id := ctx.Param("id")
		as.Equal("", id)
	})
}

func TestRequest_Close(t *testing.T) {
	var r = &Request{Header: NewHttpHeader(http.Header{}), Body: bytes.NewBufferString("")}
	r.Close()
	assert.Nil(t, r.Body)
	assert.Nil(t, r.Header)
}
