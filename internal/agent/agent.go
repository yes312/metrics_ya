package agent

import (
	"context"
	"sync"
	"time"

	metr "github.com/yes312/metrics/internal/agent/metrics"
	"github.com/yes312/metrics/internal/utils"
	"go.uber.org/zap"
)

type Agent struct {
	config *Config
	Logger *zap.SugaredLogger
}

func New(config *Config) *Agent {
	return &Agent{
		config: config,
	}
}

func (a *Agent) Start(ctx context.Context, wg *sync.WaitGroup) error {

	loger, err := utils.NewLogger("Info")
	if err != nil {
		return err
	}
	a.Logger = loger

	in := make(chan []metr.Metrics)
	out := make(chan []metr.Metrics)

	defer close(in)
	defer close(out)

	// собираем метрики раз в PollInterval и помещаем в один канал
	wg.Add(1)
	go collect(ctx, in, a.config.PollInterval, wg)

	// отправляем воркерам
	wg.Add(1)
	go sendWithInterval(ctx, in, out, a.config.ReportInterval, wg)

	// Запускаем пул воркеров в количестве RateLimit
	for i := 0; i < a.config.RateLimit; i++ {
		wg.Add(1)
		go worker(ctx, a, out, wg)
	}
	wg.Wait()
	a.Logger.Info("Завершение работы агента.")
	return nil
}

// возможно стоит функции расположенные ниже поместить в отдельный пакет, но пока не могу придумать для ниего название
func collect(ctx context.Context, in chan<- []metr.Metrics, t int, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(t) * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			in <- metr.GetMetrics()
		}
	}
}

func sendWithInterval(ctx context.Context, in <-chan []metr.Metrics, out chan<- []metr.Metrics, t int, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(t) * time.Second)
	var metr []metr.Metrics
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			out <- metr
		case metr = <-in:
		}
	}
}

func worker(ctx context.Context, a *Agent, out <-chan []metr.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case metr := <-out:
			err := a.makeChuncsRequestJSON(ctx, metr)
			if err != nil {
				a.Logger.Error(err)
			}
		}
	}
}
