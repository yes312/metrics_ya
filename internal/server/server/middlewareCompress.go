package server

import (
	"net/http"
	"strings"

	"github.com/yes312/metrics/internal/utils"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)

		} else {
			w.Header().Set("Content-Type", "text/html")
		}

		ow := w

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {

			cr, err := utils.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		var supportGzip bool
		values := r.Header.Values("Accept-Encoding")
		for _, v := range values {
			if strings.Contains(v, "gzip") {
				supportGzip = true
				continue
			}
		}
		if supportGzip {

			cw := utils.NewCompressWriter(w)
			cw.Header().Set("Content-Encoding", "gzip")

			ow = cw
			defer cw.Close()

		}

		next.ServeHTTP(ow, r)

	})

}

var _ http.ResponseWriter = &utils.CompressWriter{}
