package agent

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	metr "github.com/yes312/metrics/internal/agent/metrics"
	"github.com/yes312/metrics/internal/utils"
)

var pauseDurations = []int{0, 1, 3, 5}

func (a *Agent) makeChuncsRequestJSON(ctx context.Context, metrics []metr.Metrics) error {
	// тут установим размер пакета
	var chuncSize = 40

	u := &url.URL{
		Scheme: "http",
		Host:   a.config.DestinationAdress,
		Path:   "/updates",
	}

	client := resty.New()
	gzWriter := utils.NewGzWriter()

	chuncs := prepareChuncs(metrics, chuncSize)

	for _, chunc := range chuncs {

		body, err := json.Marshal(chunc)
		if err != nil {
			a.Logger.Info("Ошибка маршалинга", err)
		}

		body, err = gzWriter.CompressData(body)
		if err != nil {
			a.Logger.Info(err)
		}

		if a.config.Key != "" {

			signHead := utils.GetSign(body, a.config.Key)
			client.SetHeader("HashSHA256", signHead)

		}

		for _, pause := range pauseDurations {
			time.Sleep(time.Duration(pause) * time.Second)

			_, err = client.R().
				SetContext(ctx).
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept-Encoding", "gzip").
				SetHeader("Content-Encoding", "gzip").
				SetBody(body).
				Post(u.Redacted())

			if err != nil {
				if !utils.OnDialErr(err) {
					a.Logger.Error("НЕвостановимая ошибка ", err)
				}
				a.Logger.Error("восстановимая ошибка ", err)
			} else {
				break
			}

		}
	}
	return nil
}

func prepareChuncs(slice []metr.Metrics, n int) [][]metr.Metrics {
	var chunks [][]metr.Metrics
	for i := 0; i < len(slice); i += n {
		end := i + n

		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
