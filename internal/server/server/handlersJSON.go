package server

import (
	"encoding/json"
	"io"
	"net/http"

	m "github.com/yes312/metrics/internal/server/storage"
)

func (s *Server) updateJSON(w http.ResponseWriter, r *http.Request) {

	var me m.Metrics
	body, err := io.ReadAll(r.Body)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
		s.logger.Error(err)
		return
	}

	err = json.Unmarshal(body, &me)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
		s.logger.Error(err)
		return
	}

	switch me.MType {
	case "counter":
		s.storage.UpdateCounter(r.Context(), me.ID, *me.Delta)
	case "gauge":
		s.storage.UpdateGauge(r.Context(), me.ID, *me.Value)
	}

	sMetr, err := s.storage.GetMetric(r.Context(), me.MType, me.ID)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusNotFound)
		s.logger.Error("Ошибка получения метрики из хранилища сразу после записи в него", err)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(sMetr)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusNotFound)
		s.logger.Error(err)
		return
	}
	setResponseHeaders(w, ApplicationJSON, http.StatusOK)

}

func (s *Server) valueJSON(w http.ResponseWriter, r *http.Request) {

	var me m.Metrics
	body, err := io.ReadAll(r.Body)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
		s.logger.Errorln("Ошибка чтения из Body", err)
		return
	}
	err = json.Unmarshal(body, &me)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
		s.logger.Errorln("Ошибка Unmarshal ", err)
		return
	}

	sMetr, err := s.storage.GetMetric(r.Context(), me.MType, me.ID)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusNotFound)
		s.logger.Errorln("Ошибка получения метрики из хранилища", err, me.MType, me.ID)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(sMetr)
	if err != nil {
		setResponseHeaders(w, ApplicationJSON, http.StatusNotFound)
		s.logger.Errorln("Ошибка Marshal ", err)
		return
	}
	setResponseHeaders(w, ApplicationJSON, http.StatusOK)

}

func (s *Server) updates(w http.ResponseWriter, r *http.Request) {

	var me []m.Metrics
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Ошибка чтения body(updates)", err)
		setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &me)
	if err != nil {
		s.logger.Error("Ошибка маршалинга(updates)", err)
		setResponseHeaders(w, ApplicationJSON, http.StatusBadRequest)
		return
	}

	// for _, v := range me {
	// 	if v.MType == "gauge" {
	// 		fmt.Println(v.ID, v.MType, *v.Value)
	// 	} else {
	// 		fmt.Println(v.ID, v.MType, *v.Delta)
	// 	}
	// }

	s.storage.UpdateAllMetrics(r.Context(), &me)
	setResponseHeaders(w, ApplicationJSON, http.StatusOK)

}
