package metrics

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var m = &runtime.MemStats{}

type MemStats struct {
	ID   string
	metr float64
}

var c = counter()

func GetMemStatsMetrics() *[]Metrics {

	runtime.ReadMemStats(m)
	var metrics = []MemStats{
		{ID: "Alloc", metr: float64(m.Alloc)},
		{ID: "BuckHashSys", metr: float64(m.BuckHashSys)},
		{ID: "Frees", metr: float64(m.Frees)},
		{ID: "GCCPUFraction", metr: m.GCCPUFraction},
		{ID: "GCSys", metr: float64(m.GCSys)},
		{ID: "HeapAlloc", metr: float64(m.HeapAlloc)},
		{ID: "HeapIdle", metr: float64(m.HeapIdle)},
		{ID: "HeapInuse", metr: float64(m.HeapInuse)},
		{ID: "HeapObjects", metr: float64(m.HeapObjects)},
		{ID: "HeapReleased", metr: float64(m.HeapReleased)},
		{ID: "HeapSys", metr: float64(m.HeapSys)},
		{ID: "LastGC", metr: float64(m.LastGC)},
		{ID: "Lookups", metr: float64(m.Lookups)},
		{ID: "MCacheInuse", metr: float64(m.MCacheInuse)},
		{ID: "MCacheSys", metr: float64(m.MCacheSys)},
		{ID: "MSpanInuse", metr: float64(m.MSpanInuse)},
		{ID: "MSpanSys", metr: float64(m.MSpanSys)},
		{ID: "Mallocs", metr: float64(m.Mallocs)},
		{ID: "NextGC", metr: float64(m.NextGC)},
		{ID: "NumForcedGC", metr: float64(m.NumForcedGC)},
		{ID: "NumGC", metr: float64(m.NumGC)},
		{ID: "OtherSys", metr: float64(m.OtherSys)},
		{ID: "PauseTotalNs", metr: float64(m.PauseTotalNs)},
		{ID: "StackInuse", metr: float64(m.StackInuse)},
		{ID: "StackSys", metr: float64(m.StackSys)},
		{ID: "Sys", metr: float64(m.Sys)},
		{ID: "TotalAlloc", metr: float64(m.TotalAlloc)},
	}

	metr := GetGaugeMetrics(&metrics)
	r := GetNewGaugeMetric("RandomValue", float64(rand.Float64()))
	*metr = append(*metr, r)
	c := GetNewCountMetric("PollCount")
	*metr = append(*metr, c)

	return metr
}

func generator(f func() *[]Metrics, inputCh chan []Metrics) chan []Metrics {

	metrics := f()
	go func() {
		defer close(inputCh)
		inputCh <- *metrics
	}()

	return inputCh

}

func GetMetrics() []Metrics {
// создаем два канала, в них в generator кладем слайсы метрик и закрываем каналы. 
// mergeChannels  ждет данные из каналов и сливает в один канал
	ch1 := make(chan []Metrics)
	ch2 := make(chan []Metrics)

	go func() {
		generator(GetMemStatsMetrics, ch1)
	}()
	go func() {
		generator(GetGopsutilMetrics, ch2)
	}()

	var a []Metrics
	for n := range mergeChannels(ch1, ch2) {
		a = append(a, n...)
	}

	return a
}

func mergeChannels(c1, c2 chan []Metrics) chan []Metrics {

	var wg sync.WaitGroup
	out := make(chan []Metrics, 2)
	defer close(out)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for n := range c1 {
			out <- n
		}
	}()
	go func() {
		defer wg.Done()
		for n := range c2 {
			out <- n
		}
	}()

	wg.Wait()

	return out
}

func GetGopsutilMetrics() *[]Metrics {

	m := make([]Metrics, 0, 3)

	v, _ := mem.VirtualMemory()

	cpu, _ := cpu.Percent(0, true)

	m = append(m, GetNewGaugeMetric("TotalMemory", float64(v.Total)))
	m = append(m, GetNewGaugeMetric("FreeMemory", float64(v.Free)))
	m = append(m, GetNewGaugeMetric("CPUutilization1", float64(cpu[0])))

	return &m

}

func GetNewCountMetric(id string) Metrics {

	m := Metrics{
		ID:    id,
		MType: "count",
		Delta: new(int64),
		Value: new(float64),
	}
	*m.Delta = c()

	return m

}

func GetNewGaugeMetric(id string, value float64) Metrics {

	m := Metrics{
		ID:    id,
		MType: "gauge",
		Delta: new(int64),
		Value: new(float64),
	}
	*m.Value = float64(value)

	return m

}

func GetGaugeMetrics(mem *[]MemStats) *[]Metrics {

	metrics := make([]Metrics, 0, len(*mem))

	for _, v := range *mem {
		metrics = append(metrics, GetNewGaugeMetric(v.ID, v.metr))
	}

	return &metrics

}

// замыкание-счетчик
func counter() func() int64 {
	count := int64(0)
	return func() int64 {
		count++
		return count
	}
}
