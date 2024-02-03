// import (
// 	"testing"

// 	storage "github.com/yes312/metrics/internal/server/storage/internalStorage"
// )

//	func TestMetricsFileManager_WriteMetr(t *testing.T) {
//		type args struct {
//			metrics *[]storage.Metrics
//		}
//		tests := []struct {
//			name    string
//			manager *MetricsFileManager
//			args    args
//			wantErr bool
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				if err := tt.manager.WriteMetr(tt.args.metrics); (err != nil) != tt.wantErr {
//					t.Errorf("MetricsFileManager.WriteMetr() error = %v, wantErr %v", err, tt.wantErr)
//				}
//			})
//		}
//	}
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

type CaseComp1 struct {
	method             string
	name               int
	url                string
	incMetr            m.Metrics
	expectedStatusCode int
	expectedBody       m.Metrics
}

type Suite struct {
	suite.Suite
	cases  []CaseComp1
	server *httptest.Server
}

func TestSuite1(t *testing.T) {
	suite.Run(t, &Suite{})
}

func (ts *Suite) SetupTest() {

	ts.T().Log("Setup test parameters")

	tCase := CaseComp1{method: "POST", name: 1, url: "/update", expectedStatusCode: http.StatusOK,
		expectedBody: m.Metrics{ID: "myMetr", MType: "gauge", Value: new(float64)},
		incMetr:      m.Metrics{ID: "myMetr", MType: "gauge", Value: new(float64)},
	}
	*tCase.expectedBody.Value = float64(99999.111)
	*tCase.incMetr.Value = float64(99999.111)
	ts.cases = append(ts.cases, tCase)

}

func (ts *Suite) SetupSuite() {

	config, _ := NewConfig(Flags{
		A: "localhost:8080",
		F: "t.json",
		I: 0,
		R: true,
	})
	ctx := context.Background()
	server := New(config)
	loger, err := utils.NewLogger(server.config.LoggerLevel)
	if err != nil {
		ts.T().Fatal(err)
	}
	server.logger = loger

	server.configureMux()
	server.storage = memstorage.NewMemStorage()
	server.FileStorageInit(ctx)
	server.StartSavingMetrToFile(ctx)

	ts.server = httptest.NewServer(server.mux)

}

func (ts *Suite) TearDownSuite() {
	ts.T().Log("TearDownSuite")
	ts.server.Close()

}

func (ts *Suite) TestCompress1() {
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
			// SetDoNotParseResponse(true).
			Post(ts.server.URL + tt.url)

		ts.Equal(tt.expectedStatusCode, resp.StatusCode())
		ts.NoError(err)

		fmt.Println("Response body:", resp.Body())
		fmt.Println("Response status:", resp.Status())
		fmt.Println("Response body:", string(resp.Body()))

		var fromResp m.Metrics

		err = json.Unmarshal(resp.Body(), &fromResp)
		ts.NoError(err)
		fmt.Println("Response body after Unmarsh:", fromResp)
		ts.Equal(tt.expectedBody, fromResp)
	}
	// ts.server.Close()
	// ts.SetupSuite()

}
