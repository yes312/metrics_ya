package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
)

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	var p []byte
	_, err := r.Read(p)

	if err != nil {
		return nil, fmt.Errorf("ошибка  чтения gzip.NewReader %w", err)
	}

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("ошибка  создания gzip.NewReader %v", err)
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {

	gzipw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		log.Println(fmt.Errorf("ошибка создания gzip.Writer %w", err))
	}
	cw := &CompressWriter{
		w:  w,
		zw: gzipw,
	}
	return cw
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {

	return c.zw.Write(p)
	// return c.w.Write(b)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *CompressWriter) Close() error {
	return c.zw.Close()
}

func Decompress(in []byte) []byte {

	gz, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		log.Println(err)
	}
	defer gz.Close()

	b, err := io.ReadAll(gz)
	if err != nil {
		log.Println(err)
	}
	log.Println("Decompressed", string(b))
	return b
}

func CompressGzip(body []byte) []byte {

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(body)
	if err != nil {
		log.Fatalf("Ошибка сжатия: %v", err)
	}

	if err := zw.Close(); err != nil {
		log.Fatalf("Ошибка закрытия writer: %v", err)
	}
	return buf.Bytes()
}

type GzipWriter struct {
	*gzip.Writer
}

func (w *GzipWriter) CompressData(data []byte) ([]byte, error) {

	var buf bytes.Buffer

	w.Writer.Reset(&buf)

	_, err := w.Write(data)
	if err != nil {
		// return nil, fmt.Errorf("ошибка сжатия: %v", err)
		log.Printf("ошибка сжатия: %v", err)
	}

	if err := w.Close(); err != nil {
		// return nil, fmt.Errorf("ошибка закрытия gzip.Writer: %v", err)
		log.Printf("ошибка закрытия gzip.Writer: %v", err)
	}
	w.Reset(&buf)
	return buf.Bytes(), nil
}

func NewGzWriter() *GzipWriter {

	gzWriter := GzipWriter{}
	gzWriter.Writer, _ = gzip.NewWriterLevel(nil, gzip.DefaultCompression)
	return &gzWriter

}
