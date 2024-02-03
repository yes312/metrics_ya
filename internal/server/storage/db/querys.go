package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	m "github.com/yes312/metrics/internal/server/storage"
	"github.com/yes312/metrics/internal/utils"
)

var counter = "counter"
var gauge = "gauge"
var ErrMetricNotFound = errors.New("metric not found")

func (storage *Storage) UpdateGauge(ctx context.Context, metricID string, value float64) error {

	query := `INSERT INTO metrics (id, mtype, value) 
	VALUES ($1, $2, $3)  
	ON CONFLICT (id, mtype)	 
	DO UPDATE SET value = EXCLUDED.value;`

	err := storage.ExecAttempt(ctx, query, metricID, gauge, value)

	if err != nil {
		return fmt.Errorf("ошибка UpdateGauge %w", err)
	}
	return nil
}

func (storage *Storage) UpdateCounter(ctx context.Context, metricID string, delta int64) error {

	query := `INSERT INTO metrics (id, mtype, delta)
	VALUES ($1, $2, $3)
	ON CONFLICT (id, mtype)
	DO UPDATE SET delta = metrics.delta + EXCLUDED.delta;`

	err := storage.ExecAttempt(ctx, query, metricID, counter, delta)

	if err != nil {
		return fmt.Errorf("ошибка UpdateCounter %w", err)
	}
	return nil
}

func (storage *Storage) GetMetric(ctx context.Context, metricType string, metricID string) (m.Metrics, error) {

	var me m.Metrics

	row := storage.DB.QueryRowContext(ctx, `SELECT TRIM(id), TRIM(mtype), delta, value FROM metrics WHERE id=$1;`, metricID)

	err := row.Scan(&me.ID, &me.MType, &me.Delta, &me.Value)

	if err == sql.ErrNoRows {
		return m.Metrics{}, ErrMetricNotFound
	}
	if err != nil {
		return m.Metrics{}, err
	}
	return me, nil

}

func (storage *Storage) GetAllMetrics(ctx context.Context) (*[]m.Metrics, error) {

	rows, err := storage.DB.QueryContext(ctx, `SELECT TRIM(id), TRIM(mtype), delta, value FROM metrics;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var me m.Metrics
	var metrics []m.Metrics

	for rows.Next() {
		{
			err := rows.Scan(&me.ID, &me.MType, &me.Delta, &me.Value)
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, me)
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &metrics, err
}

func (storage *Storage) UpdateAllMetrics(ctx context.Context, metrics *[]m.Metrics) error {

	var delta int64
	var value float64

	newMetr := prepareMetrics(metrics)

	builder := strings.Builder{}
	builder.WriteString("INSERT INTO metrics (id, mtype, delta, value)\n")
	builder.WriteString("VALUES\n")

	lenM := len(*newMetr) - 1
	var count int
	for m, val := range *newMetr {

		if m.MType == counter {
			delta = *val.Delta
			builder.WriteString(fmt.Sprintf("('%s', '%s', %v, %s)", m.ID, m.MType, delta, "NULL"))
		} else {
			value = *val.Value
			builder.WriteString(fmt.Sprintf("('%s', '%s', %s, %v)", m.ID, m.MType, "NULL", value))
		}

		if count == lenM {
			builder.WriteString("\n")
		} else {
			builder.WriteString(",\n")
		}
		count++
	}
	builder.WriteString("ON CONFLICT (id, mtype)\n")
	builder.WriteString("DO UPDATE SET delta = metrics.delta + EXCLUDED.delta, value = EXCLUDED.value;")
	query := builder.String()
	// В целях отладки
	// fmt.Println(query)

	err := storage.ExecAttempt(ctx, query)

	if err != nil {
		return fmt.Errorf("ошибка UpdateAllMetrics %w", err)
	}
	return nil
}

type key struct {
	ID, MType string
}
type val struct {
	Value *float64
	Delta *int64
}

// в тестах подсовывают метрики с одинаковыми ID и MType. Чтобы избавиться от дублей преобразуем метрики в мапу.
// gauge просто заменяем, counter суммируем.
// TODO: наверное можно вынести эту функцию в отдельный пакет
func prepareMetrics(metrics *[]m.Metrics) *map[key]val {

	newMetrics := make(map[key]val)

	for _, metrica := range *metrics {

		if metrica.MType == counter {
			v, ok := newMetrics[key{metrica.ID, metrica.MType}]
			if ok {
				a := *v.Delta + *metrica.Delta
				x := &a
				newMetrics[key{metrica.ID, metrica.MType}] = val{metrica.Value, x}
			} else {
				newMetrics[key{metrica.ID, metrica.MType}] = val{metrica.Value, metrica.Delta}
			}

		} else {
			newMetrics[key{metrica.ID, metrica.MType}] = val{metrica.Value, metrica.Delta}
		}
	}

	return &newMetrics
}

// ExecAttempt функция реализует ExecContext, транзакции и повторение при восстановимой ошибке
func (storage *Storage) ExecAttempt(ctx context.Context, query string, args ...any) error {

	pauseDurations := []int{0, 1, 3, 5}

	for _, pause := range pauseDurations {

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(pause) * time.Second):
		}

		tx, err := storage.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return fmt.Errorf("ошибка при создании транзакции %w", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)

		if err != nil {
			if !utils.OnDialErr(err) {
				return fmt.Errorf("НЕвостановимая ошибка %w", err)
			}
			storage.logger.Info("восстановимая ошибка %v", err)
		} else {
			tx.Commit()
			break
		}

	}

	return nil

}
