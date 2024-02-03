package utils

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var ErrorWrongURLFlag = errors.New("wrong url flag")

func GetValidURL(urlString string) (string, error) {

	if !strings.HasPrefix(urlString, "http") {
		urlString = "http://" + urlString
	}
	u, err := url.Parse(urlString)
	if err != nil {
		return "", ErrorWrongURLFlag
	}

	if !u.IsAbs() || u.Port() == "" || u.Hostname() == "" {
		return "", ErrorWrongURLFlag
	}
	return u.Host, nil

}

func OnDialErr(err error) bool {
	var oe *net.OpError
	if errors.As(err, &oe) {
		return oe.Op == "dial"
	}
	return false
}
