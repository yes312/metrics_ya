package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricsSlise(t *testing.T) {

	m := GetMemStatsMetrics()
	assert.IsType(t, &[]Metrics{}, m, "m должен быть указателем на *[]Metrics")

}
