package server

import (
	"net/http"

	m "github.com/yes312/metrics/internal/server/storage"
)

type Metrics m.Metrics

// используем middlevare для записи файла, когда интервал обновления равен 0
// запись будет при каждом вызове updateJSON
func (s *Server) SaveMetrToFile(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
		metrics, _ := s.storage.GetAllMetrics(r.Context())
		s.metricsFileManager.WriteMetr(metrics)

	})
}
