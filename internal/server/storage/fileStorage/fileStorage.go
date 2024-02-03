package filestorage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	m "github.com/yes312/metrics/internal/server/storage"
)

// type Metrics m.Metrics

type MetricsFileManager struct {
	file    *os.File
	decoder *json.Decoder
	encoder *json.Encoder
}

func NewMetricsFileManager(path string, storeInterval int) (*MetricsFileManager, error) {

	// создаем каталоги если их нет и если путь содержит каталоги
	dir, _ := filepath.Split(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &MetricsFileManager{}, fmt.Errorf("не удалось создать дериктории по пути из настроек:%v ", err)
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return &MetricsFileManager{}, fmt.Errorf("не удалось открыть или создать файл %v :%v ", path, err)
	}

	return &MetricsFileManager{file: file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
	}, nil

}

func (manager *MetricsFileManager) ReadMetr() (*[]m.Metrics, error) {
	log.Println("Чтение из файла")
	var metrics []m.Metrics
	err := manager.decoder.Decode(&metrics)
	defer manager.file.Sync() //111111111111

	return &metrics, err
}

func (manager *MetricsFileManager) WriteMetr(metrics *[]m.Metrics) error {
	log.Println("Запись в файл")
	// обрезаем файл до нуля и переносим курсор на начало
	manager.file.Truncate(0)
	manager.file.Seek(0, 0)
	err := manager.encoder.Encode(metrics)

	defer manager.file.Sync()
	return err

}

func (manager *MetricsFileManager) Close() error {

	return manager.file.Close()
}
