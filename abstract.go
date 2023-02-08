package uRouter

import (
	"bytes"
	"io"
)

type (
	LoggerInterface interface {
		Debug(format string, v ...interface{})
		Info(format string, v ...interface{})
		Warn(format string, v ...interface{})
		Error(format string, v ...interface{})
		Panic(format string, v ...interface{})
	}

	BufferPoolInterface interface {
		SetSize(size int)
		Get() *bytes.Buffer
		Put(b *bytes.Buffer)
	}

	Header interface {
		Set(key, value string)
		Get(key string) string
		Del(key string)
		Len() int
		Range(f func(key, value string))
	}

	BytesReader interface {
		io.Reader
		Bytes() []byte
	}

	Closer interface {
		Close()
	}

	Encoder interface {
		Encode(v interface{}) error
	}

	Decoder interface {
		Decode(v interface{}) error
	}

	Codec interface {
		NewEncoder(w io.Writer) Encoder
		NewDecoder(r io.Reader) Decoder
		Encode(v interface{}) ([]byte, error)
		Decode(data []byte, v interface{}) error
	}
)
