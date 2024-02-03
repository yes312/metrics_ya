package server

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yes312/metrics/internal/utils"
)

func TestNewConfigPositive(t *testing.T) {

	tests := []struct {
		name     string
		arg      Flags
		wantConf *Config
	}{
		{name: "###1 +", arg: Flags{A: "127.0.0.1:8080"}, wantConf: &Config{NetworkAdress: "127.0.0.1:8080", LoggerLevel: "Info"}},
		{name: "###2 +", arg: Flags{A: "127.0.0.1:800"}, wantConf: &Config{NetworkAdress: "127.0.0.1:800", LoggerLevel: "Info"}},
		{name: "###3 +", arg: Flags{A: "localhost:800"}, wantConf: &Config{NetworkAdress: "localhost:800", LoggerLevel: "Info"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			haveConf, err := NewConfig(tt.arg)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantConf, haveConf)
		})
	}
}

// проверяем неверно неверное сетевое имя
func TestNewConfigErr(t *testing.T) {

	tests := []struct {
		name     string
		arg      Flags
		wantConf *Config
		wantErr  error
	}{
		{name: "###1 +", arg: Flags{A: "127.0.0.1"}, wantErr: utils.ErrorWrongURLFlag},
		{name: "###2 +", arg: Flags{A: "Hello"}, wantErr: utils.ErrorWrongURLFlag},
		{name: "###3 +", arg: Flags{A: "localhost"}, wantErr: utils.ErrorWrongURLFlag},
		{name: "###4 +", arg: Flags{A: ":8080"}, wantErr: utils.ErrorWrongURLFlag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := NewConfig(tt.arg)
			assert.Error(t, err, tt.wantErr)

		})
	}
}

// проверка чтения из  из env
func TestServerConfigEnv(t *testing.T) {

	tests := []struct {
		name     string
		arg      Flags
		wantConf *Config
	}{
		{name: "###1 +", arg: Flags{A: "127.0.0.1:80"},
			wantConf: &Config{
				NetworkAdress:   "127.0.0.1:8080",
				LoggerLevel:     "Info",
				StoreInterval:   100,
				FileStoragePass: "/tmp/123.txt",
				Restore:         false,
				DBAdress:        "123",
				Key:             "secret",
			}},
	}

	environments := newEnwPropertys()
	environments.Setenv()
	defer environments.DelEnv()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			haveConf, err := NewConfig(tt.arg)
			fmt.Println(haveConf, tt.wantConf)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantConf, haveConf)

		})
	}
	// DelEnv(envs)
}

type envProperty struct {
	name  string
	value string
}
type envs []envProperty

func (e *envs) Setenv() {

	for _, v := range *e {
		err := os.Setenv(v.name, v.value)
		if err != nil {
			log.Fatal("Ошибка при установке переменной окружения:", err)
		}
	}

}

func (e *envs) DelEnv() {

	for _, v := range *e {

		err := os.Unsetenv(v.name)
		if err != nil {
			fmt.Println("Ошибка при удалении переменной окружения:", v.name, err)
		}

	}
}

func newEnwPropertys() *envs {

	return &envs{{"ADDRESS", "127.0.0.1:8080"},
		{"STORE_INTERVAL", "100"},
		{"FILE_STORAGE_PATH", "/tmp/123.txt"},
		{"RESTORE", "false"},
		{"DATABASE_DSN", "123"},
		{"KEY", "secret"},
	}

}
