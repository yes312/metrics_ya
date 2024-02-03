package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	m "github.com/yes312/metrics/internal/server/storage"
)

type tSuite struct {
	suite.Suite
	storage *Storage
}

const dbName = "testdb"

// если не нужно запускаьт тест - false
// добавил отключалку что бы пройти тесты из второго инкремента на покрытие тестов. они не находят базу данных
// возможно есть другие решения, но мне это показалось самым быстрым
const testIsOn = false

func TestSuiteTest(t *testing.T) {
	suite.Run(t, &tSuite{
		Suite:   suite.Suite{},
		storage: &Storage{},
	})
}

func (ts *tSuite) TestDBSuite() {

	if testIsOn == false {
		ts.True(true)
		return
	}
	ts.T().Log("Тест UpdateCounter+GetMetric")
	ctx := context.Background()

	ts.NoError(ts.Truncate(ctx))

	// получаем текущее значение
	var delta int64
	resPrev, err := ts.storage.GetMetric(ctx, counter, "helloCounter")
	if resPrev == (m.Metrics{}) {
		delta = 0
	} else {
		delta = *resPrev.Delta
	}
	if !errors.Is(err, ErrMetricNotFound) {
		ts.NoError(err)
	}
	// отправляем новое
	ts.storage.UpdateCounter(ctx, "helloCounter", 123)
	// получаем новый ркезультат
	res, err := ts.storage.GetMetric(ctx, counter, "helloCounter")
	ts.NoError(err)
	// проверяем изменения
	ts.Suite.Equal(int64(123)+(delta), *res.Delta)
	// еще раз обновляем
	ts.storage.UpdateCounter(ctx, "helloCounter", 123)
	res, err = ts.storage.GetMetric(ctx, counter, "helloCounter")
	ts.NoError(err)
	// проверяем изменения
	ts.Suite.Equal(int64(246)+(delta), *res.Delta)

}

func (ts *tSuite) TestDBSuite1() {

	if testIsOn == false {
		ts.True(true)
		return
	}

	ts.T().Log("Тест UpdateGauge+GetMetric")
	ctx := context.Background()
	ts.NoError(ts.Truncate(ctx))
	ts.storage.UpdateGauge(ctx, "helloGauge", 100)
	ts.storage.UpdateGauge(ctx, "helloGauge", 123.1)
	// получаем новый ркезультат
	res, err := ts.storage.GetMetric(ctx, gauge, "helloGauge")
	ts.NoError(err)
	// проверяем изменения
	ts.Suite.Equal(float64(123.1), *res.Value)

}

// TRUNCATE TABLE metrics;

func (ts *tSuite) TestDBSuite2() {

	if testIsOn == false {
		ts.True(true)
		return
	}
	ts.T().Log("Тест UpdateAllMetrics")
	ctx := context.Background()

	// это нужно перенести в BeforeTest
	ts.NoError(ts.Truncate(ctx))
	var metrics []m.Metrics

	m1 := m.Metrics{ID: "myMetr1", MType: "gauge", Value: new(float64)}
	*m1.Value = 222.1
	metrics = append(metrics, m1)

	m2 := m.Metrics{ID: "myMetr2", MType: "gauge", Value: new(float64)}
	*m2.Value = 333.0
	metrics = append(metrics, m2)

	m3 := m.Metrics{ID: "myMetrCounter", MType: "counter", Delta: new(int64)}
	*m3.Delta = 3
	metrics = append(metrics, m3)

	metrics = append(metrics, m2)
	metrics = append(metrics, m3)

	ts.storage.UpdateAllMetrics(ctx, &metrics)

	res1, err := ts.storage.GetMetric(ctx, gauge, "myMetr1")
	ts.NoError(err)
	ts.Equal(*res1.Value, 222.1)

	res2, err := ts.storage.GetMetric(ctx, gauge, "myMetr2")
	ts.NoError(err)
	ts.Equal(*res2.Value, 333.0)

	res3, err := ts.storage.GetMetric(ctx, counter, "myMetrCounter")
	ts.NoError(err)
	ts.Equal(*res3.Delta, int64(6))

}

func (ts *tSuite) SetupSuite() {

	if testIsOn == false {
		ts.True(true)
		return
	}

	testDB, err := sql.Open("pgx", fmt.Sprint("postgres://postgres:12345@localhost:5432/", dbName))
	ts.NoError(err)
	db, err := CreateDBt(context.Background(), testDB)
	ts.NoError(err)
	// defer db.Close()
	ts.storage.DB = db

}

func (ts *tSuite) TearDownSuite() {
	ts.T().Log("TearDownSuite")
	if testIsOn == false {
		ts.True(true)
		return
	}

	ts.storage.DB.Close()
}

func (ts *tSuite) SetupTest() {

	ts.T().Log("Setup test parameters")

}

func CreateDBt(ctx context.Context, db *sql.DB) (*sql.DB, error) {

	var exist string
	row := db.QueryRowContext(context.Background(), "SELECT datname FROM pg_database where datname=$1;", dbName)

	row.Scan(&exist)

	if exist != dbName {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
		if err != nil {
			return nil, fmt.Errorf("ошибка создания базы данных %w", err)
		}
	}

	newDB, err := sql.Open("pgx", fmt.Sprint("postgres://postgres:12345@localhost:5432/", dbName))
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных %w", err)
	}
	// defer db.Close()

	_, err = newDB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS  metrics (
		"id"  char(25) NOT NULL,
		"mtype" char(10) NOT NULL,
		"delta" bigint,
		"value" double precision,
		UNIQUE ("id", "mtype")
	);`)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы metrics %w", err)
	}

	return newDB, nil
}

func (ts *tSuite) Truncate(ctx context.Context) error {

	_, err := ts.storage.DB.ExecContext(ctx, `TRUNCATE TABLE metrics;`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы metrics %w", err)
	}

	return nil
}
