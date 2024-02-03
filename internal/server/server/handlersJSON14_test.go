package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	m "github.com/yes312/metrics/internal/server/storage"
	memstorage "github.com/yes312/metrics/internal/server/storage/memStorage"
	"github.com/yes312/metrics/internal/utils"
)

type newCase struct {
	method             string
	name               int
	url                string
	incMetr            m.Metrics
	expectedStatusCode int
	expectedBody       m.Metrics
}

type newSuite struct {
	suite.Suite
	cases  []newCase
	server *httptest.Server
}

func TestSuiteTestJSONHandler14(t *testing.T) {
	suite.Run(t, &newSuite{})
}

func (ts *newSuite) TestJSONHandlerUpdate14() {

	config, _ := NewConfig(Flags{A: "localhost:8080", I: 20, K: "123"})
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

	ts.T().Log("Setup test parameters")
	//-----
	tCase := newCase{method: "POST", name: 1, url: "/update", expectedStatusCode: http.StatusOK,
		expectedBody: m.Metrics{ID: "myMetr", MType: "gauge", Value: new(float64)},
		incMetr:      m.Metrics{ID: "myMetr", MType: "gauge", Value: new(float64)},
	}
	*tCase.expectedBody.Value = float64(1)
	*tCase.incMetr.Value = float64(1)
	ts.cases = append(ts.cases, tCase)

	//-----
	tCase = newCase{method: "POST", name: 2, url: "/update", expectedStatusCode: http.StatusOK,
		expectedBody: m.Metrics{ID: "myMetr", MType: "counter", Delta: new(int64)},
		incMetr:      m.Metrics{ID: "myMetr", MType: "counter", Delta: new(int64)},
	}
	*tCase.expectedBody.Delta = int64(1)
	*tCase.incMetr.Delta = int64(1)
	ts.cases = append(ts.cases, tCase)

	// -----
	tCase = newCase{method: "POST", name: 3, url: "/value", expectedStatusCode: http.StatusOK,
		expectedBody: m.Metrics{ID: "PollCount", MType: "counter", Delta: new(int64)},
		incMetr:      m.Metrics{ID: "PollCount", MType: "counter", Delta: new(int64)},
	}
	*tCase.expectedBody.Delta = int64(3)
	ts.cases = append(ts.cases, tCase)

	//-----
	tCase = newCase{method: "POST", name: 4, url: "/value", expectedStatusCode: http.StatusOK,
		expectedBody: m.Metrics{ID: "GaugeMetr", MType: "gauge", Value: new(float64)},
		incMetr:      m.Metrics{ID: "GaugeMetr", MType: "gauge", Value: new(float64)},
	}
	*tCase.expectedBody.Value = float64(4)
	ts.cases = append(ts.cases, tCase)

	for _, tt := range ts.cases {

		resp, respBody := testRequestJSON14(ts.T(), ts.server, tt)
		defer resp.Body.Close()

		ts.Equal(tt.expectedStatusCode, resp.StatusCode)

		var fromResp m.Metrics

		err := json.Unmarshal(respBody, &fromResp)
		ts.NoError(err)
		ts.Equal(tt.expectedBody, fromResp)

	}

	ts.T().Log("TearDownSuite")
	ts.server.Close()
}

func testRequestJSON14(t *testing.T, ts *httptest.Server, tt newCase) (*http.Response, []byte) {

	body, err := json.Marshal(tt.incMetr)

	require.NoError(t, err)
	req, err := http.NewRequest(tt.method, ts.URL+tt.url, bytes.NewReader(body))
	require.NoError(t, err)
	//тут мы отменяем режим компрессии данных, которы работает по умолчанию
	tr := &http.Transport{}
	// tr.DisableCompression = true

	client := &http.Client{Transport: tr}

	signHead := utils.GetSign(body, "123")
	req.Header.Set("HashSHA256", string(signHead))

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// for name, values := range resp.Header {
	// 	fmt.Printf("%s: %v\n", name, values)
	// }

	sign := resp.Header.Get("HashSHA256")
	require.NotZero(t, sign)

	return resp, respBody
}

//=============================================================================================
