package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metr "github.com/yes312/metrics/internal/agent/metrics"
	"github.com/yes312/metrics/internal/utils"
)

func TestNewMetricsSlise(t *testing.T) {

	var wg sync.WaitGroup

	in := make(chan []metr.Metrics)
	defer close(in)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go collect(ctx, in, 2, &wg)
	assert.Equal(t, 32, len(<-in))
	cancel()
	wg.Wait()

}

func TestNewAgentConfigPositive(t *testing.T) {

	tests := []struct {
		name     string
		args     Flags
		wantConf *Config
	}{
		{name: "###1 +", args: Flags{A: "127.0.0.1:8080", P: 2, R: 10}, wantConf: &Config{DestinationAdress: "127.0.0.1:8080", PollInterval: 2, ReportInterval: 10}},
		{name: "###2 +", args: Flags{A: "127.0.0.1:800", P: 5, R: 12}, wantConf: &Config{DestinationAdress: "127.0.0.1:800", PollInterval: 5, ReportInterval: 12}},
		{name: "###3 +", args: Flags{A: "localhost:800", P: 5, R: 12}, wantConf: &Config{DestinationAdress: "localhost:800", PollInterval: 5, ReportInterval: 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			haveConf, err := NewAgentConfig(tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantConf, haveConf)
		})
	}
}

func TestNewAgentConfigPanic(t *testing.T) {

	tests := []struct {
		name    string
		args    Flags
		wantErr error
	}{
		{name: "###1 +", args: Flags{A: "127.0.0.1", P: 2, R: 10}, wantErr: utils.ErrorWrongURLFlag},
		{name: "###2 +", args: Flags{A: "Hello", P: 5, R: 12}, wantErr: utils.ErrorWrongURLFlag},
		{name: "###3 +", args: Flags{A: "localhost", P: 5, R: 12}, wantErr: utils.ErrorWrongURLFlag},
		{name: "###4 +", args: Flags{A: ":8080", P: 5, R: 12}, wantErr: utils.ErrorWrongURLFlag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAgentConfig(tt.args)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewAgentConfigEnv(t *testing.T) {

	tests := []struct {
		name     string
		args     Flags
		wantConf *Config
	}{
		{name: "###1 +", args: Flags{A: "127.0.0.1:80", P: 3, R: 15}, wantConf: &Config{DestinationAdress: "127.0.0.1:8080", PollInterval: 2, ReportInterval: 10, Key: "secretkey"}},
	}

	environments := newEnwPropertys()
	environments.Setenv()
	defer environments.Setenv()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			haveConf, err := NewAgentConfig(tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantConf, haveConf)

		})
	}

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
		{"REPORT_INTERVAL", "10"},
		{"POLL_INTERVAL", "2"},
		{"KEY", "secretkey"},
		{"RATE_LIMIT", "0"},
	}

}

// ======================================================
//  тестируем с использованием Suite

type Case struct {
	name     string
	args     Flags
	wantConf *Config
}

type tSuite struct {
	suite.Suite
	cases []Case
	env   *envs
}

func TestSuiteNewAgentConfigEnv(t *testing.T) {
	suite.Run(t, &tSuite{})
}

func (ts *tSuite) TestNewAgentConfigEnv() {
	for _, tt := range ts.cases {
		haveConf, err := NewAgentConfig(tt.args)
		ts.NoError(err)
		ts.Equal(tt.wantConf, haveConf, tt.name)
	}
}

func (ts *tSuite) SetupSuite() {
	ts.T().Log("Create env")

	ts.env = newEnwPropertys()
	ts.env.Setenv()

}

func (ts *tSuite) TearDownSuite() {
	ts.T().Log("Remove env")
	ts.env.DelEnv()
}

func (ts *tSuite) SetupTest() {
	ts.T().Log("Setup test parameters")

	ts.cases = []Case{

		{name: "###1 +", args: Flags{A: "127.0.0.1:80", P: 3, R: 15, K: "111", L: 4}, wantConf: &Config{DestinationAdress: "127.0.0.1:8080", PollInterval: 2, ReportInterval: 10, Key: "secretkey", RateLimit: 0}},
	}

}
