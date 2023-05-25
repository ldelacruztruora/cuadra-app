// Package gzip contains functions of the compression proccess
package gzip

import (
	"compress/gzip"
	"errors"
	"io"
	"sync"
)

var (
	// Compressor is gzip compressor which implement pool
	Compressor        gzipInterface
	errPoolCompressor = errors.New("invalid pool compressor")
)

type gzipInterface interface {
	Compress(io.Writer) (io.WriteCloser, error)
	Decompress(r io.Reader) (io.Reader, error)
}

type compressor struct {
	poolCompressor   sync.Pool
	poolDecompressor sync.Pool
}

// Compress writes compressed data to w
func (c *compressor) Compress(w io.Writer) (io.WriteCloser, error) {
	z, ok := c.poolCompressor.Get().(*writer)
	if !ok {
		return nil, errPoolCompressor
	}

	z.Writer.Reset(w) // Reset writes to w

	return z, nil
}

// Decompress writes uncompressed data to w
func (c *compressor) Decompress(r io.Reader) (io.Reader, error) {
	z, inPool := c.poolDecompressor.Get().(*reader)
	if !inPool {
		newZ, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}

		return &reader{Reader: newZ, pool: &c.poolDecompressor}, nil
	}

	err := z.Reset(r)
	if err != nil {
		c.poolDecompressor.Put(z)

		return nil, err
	}

	return z, nil
}

type writer struct {
	*gzip.Writer
	pool *sync.Pool
}

type reader struct {
	*gzip.Reader
	pool *sync.Pool
}

// Read closes the writer
func (z *reader) Read(p []byte) (n int, err error) {
	n, err = z.Reader.Read(p)
	if errors.Is(err, io.EOF) {
		z.pool.Put(z)
	}

	return n, err
}

// Close closes the writer
func (z *writer) Close() error {
	defer z.pool.Put(z)
	return z.Writer.Close()
}

func init() {
	c := &compressor{}
	c.poolCompressor.New = func() interface{} {
		return &writer{Writer: gzip.NewWriter(io.Discard), pool: &c.poolCompressor}
	}
	Compressor = c
}
