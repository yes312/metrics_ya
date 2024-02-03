package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
)

var buf string
var key string

func GetSign(b []byte, key string) string {

	h := hmac.New(sha256.New, []byte(key))
	h.Write(b)
	hash := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(hash)
	//в заголовок запроса можно класть не все символы. кодируем в разрешенные
}

var _ http.ResponseWriter = &writerWithSign{}

type writerWithSign struct {
	w http.ResponseWriter
}

func (w *writerWithSign) Write(b []byte) (int, error) {

	buf = GetSign(b, key)
	w.w.Header().Set("HashSHA256", string(buf))

	return w.w.Write(b)

}

func (w *writerWithSign) WriteHeader(statusCode int) {

	w.w.WriteHeader(statusCode)

}

func (w *writerWithSign) Header() http.Header {
	return w.w.Header()
}

func NewSignWriter(w http.ResponseWriter, k string) *writerWithSign {
	key = k
	return &writerWithSign{
		w: w,
	}
}
