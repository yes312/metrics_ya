package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	ApplicationJSON = "application/json"
	TextPlain       = "text/plain"
	TextHTML        = "text/html"
)

func setResponseHeaders(w http.ResponseWriter, contentType string, statusCode int) {

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

}

func (s *Server) GetGetAllMetrics(w http.ResponseWriter, r *http.Request) {

	metrics, _ := s.storage.GetAllMetrics(r.Context())
	log.Println(metrics)
	html := "<html><body>"
	for _, v := range *metrics {
		if v.MType == "gauge" {
			html += fmt.Sprintf("[%s] %s: %v", v.MType, v.ID, *v.Value) + "<br>"
		} else {
			html += fmt.Sprintf("[%s] %s: %v", v.MType, v.ID, *v.Delta) + "<br>"
		}
	}
	html += "</body></html>"

	fmt.Fprint(w, html)
	setResponseHeaders(w, TextHTML, http.StatusOK)

}

func (s *Server) GetMetricValue(w http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	m, err := s.storage.GetMetric(r.Context(), metricType, metricName)
	if err != nil {
		setResponseHeaders(w, TextPlain, http.StatusNotFound)
		s.logger.Error(err)
		return
	}

	if m.MType == "gauge" {
		fmt.Fprint(w, *m.Value)
	} else {
		fmt.Fprint(w, *m.Delta)
	}

	setResponseHeaders(w, TextPlain, http.StatusOK)
}

func (s *Server) gauge(w http.ResponseWriter, r *http.Request) {

	metricValue := chi.URLParam(r, "value")
	metricName := chi.URLParam(r, "name")

	value, err := strconv.ParseFloat(metricValue, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("wrong convert to float64 ", err)
		return
	}

	s.storage.UpdateGauge(r.Context(), metricName, value)
	setResponseHeaders(w, TextPlain, http.StatusOK)

}

func (s *Server) counter(w http.ResponseWriter, r *http.Request) {

	metricValue := chi.URLParam(r, "value")
	metricName := chi.URLParam(r, "name")

	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("wrong convert to int64 ", err)
		return
	}
	s.storage.UpdateCounter(r.Context(), metricName, value)
	setResponseHeaders(w, TextPlain, http.StatusOK)

}

func (s *Server) incorrectType(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	s.logger.Error("wrong metric type", r.RequestURI)
}

func (s *Server) wrongMetricName(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	s.logger.Error("wrong metric name", r.RequestURI)
}

func (s *Server) wrongMetricVolume(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	s.logger.Error("wrong metric volume", r.RequestURI)
}

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := s.storage.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("wrong db Ping", r.RequestURI, err)
	}
	setResponseHeaders(w, TextPlain, http.StatusOK)

}
