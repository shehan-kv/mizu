package query

import (
	"errors"
	"net/http"
	"strconv"
)

func ExtractString(r *http.Request, key string) *string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	return &val
}

func ExtractInt(r *http.Request, key string, defaultVal int) (int, error) {
	str := r.URL.Query().Get(key)
	if str == "" {
		return defaultVal, nil
	}
	val, err := strconv.Atoi(str)
	if err != nil || val <= 0 {
		return 0, errors.New("invalid value for " + key)
	}
	return val, nil
}

func ExtractBool(r *http.Request, key string) *bool {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	b := val == "true"
	return &b
}
