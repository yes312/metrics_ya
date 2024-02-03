package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	memstorage "github.com/yes312/metrics/internal/server/storage/memStorage"
	"github.com/yes312/metrics/internal/utils"

	"github.com/stretchr/testify/require"
)

func TestCountHandler(t *testing.T) {

	config, _ := NewConfig(Flags{A: "localhost:8080", I: 20})
	server := New(config)
	loger, err := utils.NewLogger(server.config.LoggerLevel)
	if err != nil {
		t.Fatal(err)
	}
	server.logger = loger

	server.configureMux()
	ctx := context.Background()
	server.storage = memstorage.NewMemStorage()
	server.storage.UpdateCounter(ctx, "PollCount", 2)
	server.storage.UpdateGauge(ctx, "GaugeMetr", 1)

	testServer := httptest.NewServer(server.mux)
	defer testServer.Close()

	testCases := []struct {
		method             string
		name               string
		url                string
		expectedStatusCode int
		expectedBody       string
	}{
		{method: "GET", name: "/value/counter/PollCount ", url: "/value/counter/PollCount", expectedStatusCode: http.StatusOK, expectedBody: "2"},
		{method: "GET", name: "/value/counter/GaugeMetr ", url: "/value/gauge/GaugeMetr", expectedStatusCode: http.StatusOK, expectedBody: "1"},
		{method: "GET", name: "/value/counter/uncnown ", url: "/value/counter/uncnown", expectedStatusCode: http.StatusNotFound, expectedBody: ""},
		{method: "GET", name: "/Get all metrics ", url: "/", expectedStatusCode: http.StatusOK, expectedBody: "<html><body>[gauge] GaugeMetr: 1<br>[counter] PollCount: 2<br></body></html>"},
		{method: "POST", name: "###1/update/ -", url: "/update", expectedStatusCode: http.StatusBadRequest, expectedBody: ""},
		{method: "POST", name: "###2/update/counter/ -", url: "/update/counter/", expectedStatusCode: http.StatusNotFound, expectedBody: ""},
		{method: "POST", name: "###3/update/uncnown/ -", url: "/update/uncnown/", expectedStatusCode: http.StatusBadRequest, expectedBody: ""},
		{method: "POST", name: "###4/update/counter/metr/12 +", url: "/update/counter/qwe/12", expectedStatusCode: http.StatusOK, expectedBody: ""},
		{method: "POST", name: "###5/update/counter/metr/122rt ", url: "/update/counter/metrw/122rt", expectedStatusCode: http.StatusBadRequest, expectedBody: ""},
		{method: "POST", name: "/update/gauge/1 -", url: "/update/gauge/1", expectedStatusCode: http.StatusBadRequest, expectedBody: ""},
		{method: "POST", name: "/update/gauge/metr/1 +", url: "/update/gauge/metrw/12", expectedStatusCode: http.StatusOK, expectedBody: ""},
		{method: "POST", name: "/update/gauge/metr/122rt ", url: "/update/gauge/metrw/122rt", expectedStatusCode: http.StatusBadRequest, expectedBody: ""},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			resp, respBody := testRequest(t, testServer, tt.method, tt.url)
			defer resp.Body.Close()
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, respBody)
		})

	}

}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {

	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	tr := &http.Transport{}
	tr.DisableCompression = true
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
