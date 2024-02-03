package server

import (
	"net/http"
	"time"
)

type responseData struct {
	status int
	size   int
}

type writerWithLog struct {
	http.ResponseWriter
	responseData
}

var _ http.ResponseWriter = &writerWithLog{}

func (w *writerWithLog) Write(b []byte) (int, error) {

	size, err := w.ResponseWriter.Write(b)

	w.responseData.size = size
	return size, err

}

func (w *writerWithLog) WriteHeader(statusCode int) {

	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)

}

func (s *Server) LoggerMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wLog := writerWithLog{
			w,
			responseData{},
		}

		start := time.Now()
		next.ServeHTTP(&wLog, r)
		duration := time.Since(start)

		s.logger.Infoln(
			"uri:", r.RequestURI,
			"method:", r.Method,
			"duration:", duration,
			"statusCode:", wLog.responseData.status,
			"size:", wLog.responseData.size,
		)

	})
}
