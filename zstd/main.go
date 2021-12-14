package main

import (
	"fmt"
	"io"
	"os"

	"github.com/DataDog/zstd"
)

type zstdClient struct {
	F      *os.File
	Writer *zstd.Writer
}

func newZstdClient(path string, level int) (*zstdClient, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("open file failure, nest error: %v", err)
	}

	writer := zstd.NewWriterLevel(f, level)
	return &zstdClient{f, writer}, nil
}

func (z *zstdClient) Close() error {
	if z == nil {
		return nil
	}
	z.Writer.Close()
	z.F.Close()
	return nil
}

func ZstdCompress(src string, dest string, level int) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := newZstdClient(dest, level)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w.Writer, r)
	return err
}

func ZstdDecompress(src string, dest string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, zstd.NewReader(r))
	return err
}

func main() {
	ZstdCompress("main.go", "main.go.zstd", 3)
	ZstdDecompress("main.go.zstd", "main.go.bak")
}
