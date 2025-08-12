package util

import (
	"strconv"
	"time"
)

// ToRFC3339 converts a string unix timestamp (e.g., "1234567890.123456") to RFC3339 format.
func ToRFC3339(unixTs string) (string, error) {
	if unixTs == "" {
		return "", nil
	}
	floatTs, err := strconv.ParseFloat(unixTs, 64)
	if err != nil {
		return "", err
	}
	sec := int64(floatTs)
	nsec := int64((floatTs - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec)
	return t.UTC().Format(time.RFC3339), nil
}