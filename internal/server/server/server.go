package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	db "github.com/yes312/metrics/internal/server/storage/db"
	filestorage "github.com/yes312/metrics/internal/server/storage/fileStorage"
	memstorage "github.com/yes312/metrics/internal/server/storage/memStorage"
	"github.com/yes312/metrics/internal/utils"
	"go.uber.org/zap"
)

var _ Storager = &memstorage.MemStorage{}
var _ Storager = &db.Storage{}

type Storager interface {
	db.StoragerDB
	memstorage.StoragerMem
}

type Server struct {
	server             *http.Server
	config             *Config
	mux                *chi.Mux
	storage            Storager
	metricsFileManager *filestorage.MetricsFileManager
	logger             *zap.SugaredLogger
}

func New(config *Config) *Server {

	return &Server{
		config: config,
		mux:    chi.NewRouter(),
	}
}

func (s *Server) Start(ctx context.Context) error {

	log.Println("===Запуск сервера===")
	loger, err := utils.NewLogger(s.config.LoggerLevel)
	if err != nil {
		return err
	}
	s.logger = loger
	s.configureMux()

	if s.config.DBAdress != "" {
		s.logger.Info("===Сохраняем в БД")
		config := db.NewConfig(s.config.DBAdress)
		store, err := db.New(ctx, config)
		if err != nil {
			return err
		}
		s.storage = store

	} else {
		s.logger.Info("===Сохраняем во внутреннее хранилище")

		s.storage = memstorage.NewMemStorage()
	}

	if s.config.FileStoragePass != "" {

		s.FileStorageInit(ctx)
		s.StartSavingMetrToFile(ctx)
	}

	s.server = &http.Server{
		Addr:    s.config.NetworkAdress,
		Handler: s.mux,
	}
	return s.server.ListenAndServe()
}

func (s *Server) configureMux() {

	s.mux.Route("/", func(r chi.Router) {

		r.Use(s.LoggerMW)
		if s.config.Key != "" {
			r.Use(s.CheckSign)
		}
		r.Use(GzipMiddleware)
		r.Get("/", s.GetGetAllMetrics)

		r.Get("/ping", s.Ping)

		r.Get("/value/{type}/{name}", s.GetMetricValue)

		r.Post("/value", s.valueJSON)
		r.Post("/value/", s.valueJSON)

		r.Post("/updates", s.updates)
		r.Post("/updates/", s.updates)

		r.Route("/update", func(r chi.Router) {
			// в зависимости от настроек StoreInterval настраиваем роутер
			// если интервал = 0, то добавляем middleware, который при каждом обновлении метрик
			// так же сохраняет данные в файл
			// ==============================================
			// для 11 инкремента дописываю s.config.FileStoragePass != ""
			// возможно этот middleware пригодится так же для сохранение в БД
			// но пока он сохраняет только в файл .Возможно нужно будет уточнить у ментора!!
			if s.config.StoreInterval == 0 && s.config.FileStoragePass != "" {
				r.With(s.SaveMetrToFile).Post("/", s.updateJSON)
			} else {
				r.Post("/", s.updateJSON)
			}

			r.Route("/{type}", func(r chi.Router) {
				r.Post("/", s.incorrectType)
				r.Post("/{name}", s.incorrectType)
				r.Post("/{name}/{value}", s.incorrectType)
			})
			r.Route("/counter", func(r chi.Router) {
				r.Post("/", s.wrongMetricName)
				r.Post("/{name}", s.wrongMetricVolume)
				r.Post("/{name}/{value}", s.counter)

			})
			r.Route("/gauge", func(r chi.Router) {
				r.Post("/", s.wrongMetricName)
				r.Post("/{name}", s.wrongMetricVolume)
				r.Post("/{name}/{value}", s.gauge)
			})
		})

	})
	s.logger.Info("Роутер сконфигурирован")
}

func (s *Server) FileStorageInit(ctx context.Context) {
	// если путь пустой файл не создаем и ничего не сохраняем
	if s.config.FileStoragePass != "" {
		s.logger.Info("Активируем сохраниение в файл")
		fs, err := filestorage.NewMetricsFileManager(s.config.FileStoragePass, s.config.StoreInterval)
		if err != nil {
			s.logger.Error("Ошибка при открытии файла: %v", err)
		}
		s.metricsFileManager = fs

		// пробуем или нет (в зависимости от конфига)загрузить метрики из файла
		if s.config.Restore {
			s.logger.Info("Загружаем метрики из файла")
			metrics, err := s.metricsFileManager.ReadMetr()
			if err != nil {
				s.logger.Error("Ошибка при получении метрик: %w ", err)
			} else {
				if len(*metrics) != 0 {

					s.logger.Info("Сохраняем метрики из файла в хранилище")
					s.storage.UpdateAllMetrics(ctx, metrics)
				}

			}
		}

	}

}

func (s *Server) Close() {

	if err := s.metricsFileManager.Close(); err != nil {
		s.logger.Errorf("файл %v закрылся с ошибкой %w", s.config.FileStoragePass, err)
	} else {
		s.logger.Info("Файл с метриками успешно закрыт")
	}

	if err := s.storage.Close(); err != nil {
		s.logger.Errorf("БД закрылась с ошибкой %w", err)
	} else {
		s.logger.Info("БД закрылась успешно")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorf("Ошибка при остановке http.Server", err)
	} else {
		s.logger.Info("http сервер остановлен")
	}

}

// запускает сохранение в файл с указанным интервалом
func (s *Server) StartSavingMetrToFile(ctx context.Context) {

	if s.config.FileStoragePass != "" && s.config.StoreInterval != 0 {
		go func() {
			for {
				metrics, _ := s.storage.GetAllMetrics(ctx)
				s.metricsFileManager.WriteMetr(metrics)
				time.Sleep(time.Second * time.Duration(s.config.StoreInterval))
			}

		}()
	}
}
