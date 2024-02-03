package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yes312/metrics/internal/utils"
)

func (s *Server) CheckSign(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sign := r.Header.Get("HashSHA256")

		if sign == "" {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Error("Ошибка чтения body(Sign)", err)
			setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
			return
		}

		signHead := utils.GetSign(body, s.config.Key)

		if signHead != sign {
			s.logger.Error("подпись не соответствует сообщению ")

			setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
			return

		}
		// если мы прочитали r.Body и планируем читать еще, но уже в другом месте - нужно "вернуть каретку назад".
		// это можно сделать, сохранив прочитанные данные и создав новый io.ReadCloser с этими данными, который затем можно присвоить r.Body
		r.Body = io.NopCloser(bytes.NewReader(body))

		// после того как все заголовки были записаны с помощью WriteHeader или Write, мы не можем добавить дополнительные заголовки.
		// по заданию нам нужно подписать сообщение, отправляемое с сервера
		// чтобы это обойти подменяем ResponseWriter. в нем и будем подписывать.
		//
		wSign := utils.NewSignWriter(w, s.config.Key)

		next.ServeHTTP(wSign, r)

	})
}
