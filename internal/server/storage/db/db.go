package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	m "github.com/yes312/metrics/internal/server/storage"
	"github.com/yes312/metrics/internal/utils"
	"go.uber.org/zap"
)

var _ StoragerDB = &Storage{}

type StoragerDB interface {
	UpdateGauge(context.Context, string, float64) error
	UpdateCounter(context.Context, string, int64) error
	GetMetric(context.Context, string, string) (m.Metrics, error)
	GetAllMetrics(context.Context) (*[]m.Metrics, error)
	UpdateAllMetrics(context.Context, *[]m.Metrics) error
	Ping(context.Context) error
	Close() error
}

// const dbName = "metricsdb"

type Config struct {
	DBUrl string
}

type Storage struct {
	config *Config
	DB     *sql.DB
	logger *zap.SugaredLogger
}

// EsdeathEsdeath 2 hours ago
// Не очень понятно почему именно так сделано? Почему мы просто не можем передать нам конфигурацию из main?

// когдя я это делал я не знал в каком виде придет информация о строке из теста. Я предположил что ее нужно будет парсить и потом формировать
// строку для подключения к postgre. Это было бы как раз то место где я бы это делал. но пришла готовая строка.
// лучше удалить?
func NewConfig(url string) *Config {
	return &Config{
		DBUrl: url,
	}
}

func New(ctx context.Context, config *Config) (*Storage, error) {

	logger, err := utils.NewLogger("Info")
	if err != nil {
		return nil, err
	}

	db, err := OpenDBConnection(config)
	if err != nil {
		return nil, err
	}

	db, err = CreateTable(ctx, db)
	if err != nil {
		return nil, err
	}

	// defer db.Close()

	return &Storage{
		config: config,
		DB:     db,
		logger: logger,
	}, nil
}

func OpenDBConnection(config *Config) (*sql.DB, error) {

	db, err := sql.Open("pgx", config.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных(Ping) %w", err)
	}

	return db, nil
}

func (storage *Storage) Ping(ctx context.Context) error {

	if err := storage.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("ошибка ping %w", err)
	}
	return nil
}

func (storage *Storage) Close() error {

	return storage.DB.Close()

}

func CreateTable(ctx context.Context, db *sql.DB) (*sql.DB, error) {

	// КОД НИЖЕ СОЗДАЕТ БД, ЕСЛИ ЕЕ НЕТ, НО У НАС ОНА ЕСТЬ! ОСТАВЛЮ НА ВСЯКИЙ СЛУЧАЙ
	// var exist string
	// row := db.QueryRowContext(ctx, "SELECT datname FROM pg_database where datname=$1;", dbName)
	// row.Scan(&exist)
	// if exist != dbName {
	// 	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	// 	if err != nil {
	// 		return nil, fmt.Errorf("ошибка создания базы данных %w", err)
	// 	}
	// }
	// newDB, err := sql.Open("pgx", fmt.Sprint("postgres://postgres:12345@localhost:5432/", dbName))
	// if err != nil {
	// 	return nil, fmt.Errorf("ошибка открытия базы данных %w", err)
	// }
	// defer db.Close()

	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS  metrics (
		"id"  char(25) NOT NULL,
		"mtype" char(10)NOT NULL,   
		"delta" bigint , 
		"value" double precision,
		UNIQUE ("id", "mtype")
	);`)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы metrics %w", err)
	}

	return db, nil
}
