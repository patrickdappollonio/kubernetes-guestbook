package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func allSet(key ...string) bool {
	for _, k := range key {
		if os.Getenv(k) == "" {
			return false
		}
	}
	return true
}

func envdefault(key, defval string) string {
	if s := strings.TrimSpace(os.Getenv(key)); s != "" {
		return s
	}
	return defval
}

func httpError(w http.ResponseWriter, statusCode int, format string, args ...interface{}) {
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": fmt.Sprintf(format, args...),
	})
}

func strToBytes(s string) []byte {
	return []byte(s)
}

func asJSON(w http.ResponseWriter, v interface{}) {
	json.NewEncoder(w).Encode(v)
}
