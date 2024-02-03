package agent

import (
	"errors"
	"os"
	"strconv"

	"github.com/yes312/metrics/internal/utils"
)

type Flags struct {
	A string
	R uint64
	P uint64
	K string
	L uint64
}

var ErrorWrongURLFlag = errors.New("wrong url flag")

type Config struct {
	PollInterval      int
	ReportInterval    int
	DestinationAdress string
	Key               string
	RateLimit         int
}

func NewAgentConfig(f Flags) (*Config, error) {

	c := Config{}
	if a, ok := os.LookupEnv("ADDRESS"); ok {
		c.DestinationAdress = a
	} else {
		var err error
		if c.DestinationAdress, err = utils.GetValidURL(f.A); err != nil {
			return &Config{}, err
		}
	}
	if r, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		if i, err := strconv.Atoi(r); err == nil {
			c.ReportInterval = i
		}
	} else {
		c.ReportInterval = int(f.R)
	}

	if r, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		if i, err := strconv.Atoi(r); err == nil {
			c.PollInterval = i
		}
	} else {
		c.PollInterval = int(f.P)
	}

	if k, ok := os.LookupEnv("KEY"); ok {
		if k != "" {
			c.Key = k
		} else {
			c.Key = f.K
		}
	} else {
		c.Key = f.K
	}

	if r, ok := os.LookupEnv("RATE_LIMIT"); ok {
		if i, err := strconv.Atoi(r); err == nil {
			c.RateLimit = i
		}
	} else {
		c.RateLimit = int(f.L)
	}

	return &c, nil
}
