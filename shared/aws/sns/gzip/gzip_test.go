package gzip

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompress(t *testing.T) {
	c := require.New(t)

	var buf bytes.Buffer
	cmp, err := Compressor.Compress(&buf)
	c.Nil(err)

	err = cmp.Close()
	c.Nil(err)
}

func TestDecompress(t *testing.T) {
	c := require.New(t)

	var buf bytes.Buffer
	cmp, err := Compressor.Compress(&buf)
	c.Nil(err)

	err = cmp.Close()
	c.Nil(err)

	bufReader := bytes.NewReader(buf.Bytes())
	r, err := Compressor.Decompress(bufReader)
	c.Nil(err)

	_, err = r.Read(buf.Bytes())
	c.Equal(io.EOF, err)

	_, err = Compressor.Decompress(bufReader)
	c.Equal(io.EOF, err)

	_, err = Compressor.Decompress(bufReader)
	c.Equal(io.EOF, err)

	_, err = Compressor.Decompress(bufReader)
	c.Equal(io.EOF, err)

	_, err = Compressor.Decompress(bufReader)
	c.Equal(io.EOF, err)
}

func TestMock(t *testing.T) {
	c := require.New(t)

	MockInit()

	var buf bytes.Buffer
	cmp, err := Compressor.Compress(&buf)
	c.Nil(err)

	i, err := cmp.Write([]byte{})
	c.Equal(0, i)
	c.Nil(err)

	err = cmp.Close()
	c.Nil(err)

	ForceCompressFail = true
	_, err = Compressor.Compress(&buf)
	c.Equal(ErrMock, err)

	MockInit()

	ForceWriteFail = true
	_, err = cmp.Write([]byte{})
	c.Equal(ErrMock, err)

	MockInit()

	ForceCloseFail = true
	err = cmp.Close()
	c.Equal(ErrMock, err)

	MockInit()

	r, err := Compressor.Decompress(&buf)
	c.Nil(err)

	_, err = r.Read([]byte{})
	c.Nil(err)

	ForceReadFail = true
	_, err = r.Read([]byte{})
	c.Equal(ErrMock, err)

	MockInit()

	ForceDecompressFail = true
	_, err = Compressor.Decompress(&buf)
	c.Equal(ErrMock, err)

	MockInit()

	_, err = Compressor.Decompress(&buf)
	c.Nil(err)
}

func BenchmarkTestCompress(b *testing.B) {
	c := require.New(b)
	var buf bytes.Buffer

	for i := 0; i < b.N; i++ {
		cmp, err := Compressor.Compress(&buf)
		c.Nil(err)

		err = cmp.Close()
		c.Nil(err)
	}
}
