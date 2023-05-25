package gzip

import (
	"errors"
	"io"
)

var (
	// ForceCompressFail used to make Compress fail
	ForceCompressFail = false
	// ForceDecompressFail used to make Decompress fail
	ForceDecompressFail = false
	// ForceReadFail used to make Read fail
	ForceReadFail = false
	// ForceWriteFail used to make Write fail
	ForceWriteFail = false
	// ForceCloseFail used to make Close fail
	ForceCloseFail = false
	// ErrMock error caused by mock
	ErrMock = errors.New("mock fail")
)

// MockCompressor mock used for Compressor
type MockCompressor struct {
}

// MockInit initialize mock
func MockInit() {
	Compressor = &MockCompressor{}

	// Init mock flags
	ForceCompressFail = false
	ForceDecompressFail = false
	ForceReadFail = false
	ForceWriteFail = false
	ForceCloseFail = false
}

// Compress is a mock method to test file compress
func (c *MockCompressor) Compress(w io.Writer) (io.WriteCloser, error) {
	if ForceCompressFail {
		return nil, ErrMock
	}

	return &mockWriter{}, nil
}

// Decompress is a mock method to test file decompress
func (c *MockCompressor) Decompress(r io.Reader) (io.Reader, error) {
	if ForceDecompressFail {
		return nil, ErrMock
	}

	return &mockReader{}, nil
}

type mockWriter struct {
}

type mockReader struct {
}

// Read is a mock method
func (r *mockReader) Read(p []byte) (int, error) {
	if ForceReadFail {
		return 0, ErrMock
	}

	return 0, nil
}

// Write is a mock method
func (c *mockWriter) Write(p []byte) (int, error) {
	if ForceWriteFail {
		return 0, ErrMock
	}

	return 0, nil
}

// Close is a mock method
func (c *mockWriter) Close() error {
	if ForceCloseFail {
		return ErrMock
	}

	return nil
}
