package memstorage

import (
	"context"
	"errors"
	"sync"

	m "github.com/yes312/metrics/internal/server/storage"
)

var ErrMetricNotFound = errors.New("metric not found")

var _ StoragerMem = &MemStorage{}

type StoragerMem interface {
	UpdateGauge(context.Context, string, float64) error
	UpdateCounter(context.Context, string, int64) error
	GetMetric(context.Context, string, string) (m.Metrics, error)
	GetAllMetrics(context.Context) (*[]m.Metrics, error)
	UpdateAllMetrics(context.Context, *[]m.Metrics) error
	Ping(context.Context) error
	Close() error
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
	mu      *sync.Mutex
}

func (mem *MemStorage) UpdateGauge(ctx context.Context, metricName string, value float64) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()
	mem.gauge[metricName] = value
	return nil

}

func (mem *MemStorage) UpdateCounter(ctx context.Context, metricName string, value int64) error {

	mem.mu.Lock()
	defer mem.mu.Unlock()
	mem.counter[metricName] += value
	return nil
}

func (mem *MemStorage) GetMetric(ctx context.Context, metricType, metricName string) (m.Metrics, error) {

	switch metricType {
	case "counter":
		if value, ok := mem.counter[metricName]; ok {
			return m.Metrics{ID: metricName, MType: "counter", Delta: &value}, nil
		}
		return m.Metrics{}, ErrMetricNotFound
	case "gauge":

		if value, ok := mem.gauge[metricName]; ok {
			return m.Metrics{ID: metricName, MType: "gauge", Value: &value}, nil
		}
	}
	return m.Metrics{}, ErrMetricNotFound

}

func (mem *MemStorage) GetAllMetrics(ctx context.Context) (*[]m.Metrics, error) {

	mSl := make([]m.Metrics, len(mem.gauge))
	k := 0
	for name, value := range mem.gauge {
		copyValue := value
		mSl[k] = m.Metrics{ID: name, MType: "gauge", Value: &copyValue}
		k++
	}

	for name, value := range mem.counter {
		copyValue := value
		a := m.Metrics{ID: name, MType: "counter", Delta: &copyValue}
		mSl = append(mSl, a)
	}

	return &mSl, nil
}

func NewMemStorage() *MemStorage {

	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
		mu:      &sync.Mutex{},
	}
}

func (mem *MemStorage) UpdateAllMetrics(ctx context.Context, metrics *[]m.Metrics) error {
	var err error
	for _, m := range *metrics {

		switch m.MType {
		case "counter":
			err = mem.UpdateCounter(ctx, m.ID, *m.Delta)
		case "gauge":
			err = mem.UpdateGauge(ctx, m.ID, *m.Value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *MemStorage) Close() error {
	return nil
}
func (mem *MemStorage) Ping(ctx context.Context) error {
	return nil
}
