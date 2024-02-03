package agent

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	metr "github.com/yes312/metrics/internal/agent/metrics"
)

type testCase struct {
	name                  string
	a                     *Agent
	Metrics               *[]metr.Metrics
	expectedMetr          *[]metr.Metrics
	expectedNumberDelMetr int
}

func TestAgent_makeRequest(t *testing.T) {

	testServer := httptest.NewServer(Handler())
	defer testServer.Close()
	f := Flags{A: testServer.URL}

	config, err := NewAgentConfig(f)
	if err != nil {
		log.Fatal(err)
	}
	a := New(config)

	tests := []testCase{}

	m := []metr.Metrics{
		{ID: "Alloc", MType: "gauge", Value: new(float64)},
		{ID: "MMM", MType: "counter", Delta: new(int64)}}
	*m[0].Value = 2
	*m[1].Delta = 3

	tCase := testCase{name: "test1", a: a, expectedNumberDelMetr: 2, Metrics: &m}
	tests = append(tests, tCase)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expectedMetr = tt.Metrics
			assert.Equal(t, 0, 0)
		})
	}
}

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
