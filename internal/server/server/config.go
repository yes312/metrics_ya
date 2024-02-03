package server

import (
	"log"
	"os"
	"strconv"

	"github.com/yes312/metrics/internal/utils"
)

type Flags struct {
	A string // NetworkAdress
	F string // FileStoragePass
	D string // DB conection adress
	I int    // StoreInterval
	R bool   // Resore
	K string // Key
}

type Config struct {
	NetworkAdress   string
	LoggerLevel     string
	StoreInterval   int
	FileStoragePass string
	Restore         bool
	DBAdress        string
	Key             string
}

func NewConfig(flag Flags) (*Config, error) {
	log.Println("NewConfig=================")
	c := Config{}
	if buf, ok := os.LookupEnv("ADDRESS"); ok {
		c.NetworkAdress = buf
	} else {
		var err error
		if c.NetworkAdress, err = utils.GetValidURL(flag.A); err != nil {
			return &Config{}, utils.ErrorWrongURLFlag
		}
	}
	if buf, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if interv, err := strconv.Atoi(buf); err == nil {
			c.StoreInterval = interv
		}
	} else {
		c.StoreInterval = flag.I
	}

	if buf, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.FileStoragePass = buf
	} else {
		c.FileStoragePass = flag.F
	}

	if buf, ok := os.LookupEnv("RESTORE"); ok {
		if b, err := strconv.ParseBool(buf); err == nil {
			c.Restore = b
		}
	} else {
		c.Restore = flag.R
	}

	if buf, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DBAdress = buf
	} else {
		c.DBAdress = flag.D
	}

	if k, ok := os.LookupEnv("KEY"); ok {
		if k != "" {
			c.Key = k
		} else {
			c.Key = flag.K
		}
	} else {
		c.Key = flag.K
	}

	//Задаем по умолчанию Info. Возможно стоит вынести уровень логирования в переменную окружения или сделать для нее флаг,
	//но этого не в задании
	c.LoggerLevel = "Info"

	return &c, nil

}
