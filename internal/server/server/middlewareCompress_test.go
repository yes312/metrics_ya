package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	m "github.com/yes312/metrics/internal/server/storage"
	memstorage "github.com/yes312/metrics/internal/server/storage/memStorage"
	"github.com/yes312/metrics/internal/utils"
)

type CaseComp struct {
	method             string
	name               int
	url                string
	incMetr            m.Metrics
	expectedStatusCode int
	expectedBody       m.Metrics
}

type SuiteComp struct {
	suite.Suite
	cases  []CaseComp
	server *httptest.Server
	// client *resty.Client
}

func TestSuite(t *testing.T) {
	suite.Run(t, &SuiteComp{})
}

func (ts *SuiteComp) SetupTest() {

	ts.T().Log("Setup test parameters")

	tCase := CaseComp{method: "POST", name: 1, url: "/update", expectedStatusCode: http.StatusOK,
		expectedBody: m.Metrics{ID: "myMetr", MType: "gauge", Value: new(float64)},
		incMetr:      m.Metrics{ID: "myMetr", MType: "gauge", Value: new(float64)},
	}
	*tCase.expectedBody.Value = float64(99999.111)
	*tCase.incMetr.Value = float64(99999.111)
	ts.cases = append(ts.cases, tCase)

}

func (ts *SuiteComp) SetupSuite() {

	config, _ := NewConfig(Flags{
		A: "localhost:8080",
		F: "",
		I: 300,
		R: false,
	})
	server := New(config)
	loger, err := utils.NewLogger(server.config.LoggerLevel)
	if err != nil {
		ts.T().Fatal(err)
	}
	server.logger = loger
	server.configureMux()
	ctx := context.Background()
	server.storage = memstorage.NewMemStorage()
	server.storage.UpdateCounter(ctx, "PollCount", 3)
	server.storage.UpdateGauge(ctx, "GaugeMetr", 4)

	ts.server = httptest.NewServer(server.mux)

}

func (ts *SuiteComp) TearDownSuite() {
	ts.T().Log("TearDownSuite")
	ts.server.Close()
}

func (ts *SuiteComp) TestCompress1() {
	for _, tt := range ts.cases {

		client := resty.New()

		body, err := json.Marshal(tt.incMetr)
		require.NoError(ts.T(), err)
		compBody := utils.CompressGzip(body)

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetHeader("Content-Encoding", "gzip").
			SetBody(compBody).
			Post(ts.server.URL + tt.url)

		ts.Equal(tt.expectedStatusCode, resp.StatusCode())
		ts.NoError(err)

		var fromResp m.Metrics

		err = json.Unmarshal(resp.Body(), &fromResp)
		ts.NoError(err)
		fmt.Println("Response body after Unmarsh:", fromResp)
		ts.Equal(tt.expectedBody, fromResp)
	}
}
