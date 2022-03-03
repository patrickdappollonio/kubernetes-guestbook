package storage

import (
	"strconv"
	"strings"
)

const alpha = "abcdefghijklmnopqrstuvwxyz"

func alphaOnly(s string) bool {
	for _, char := range s {
		if !strings.Contains(alpha, strings.ToLower(string(char))) {
			return false
		}
	}
	return true
}

func obfuscate(s string, length int) string {
	if len(s) > length {
		return s[:length] + strings.Repeat("*", len(s)-length) + " (len: " + strconv.Itoa(len(s)) + ")"
	}

	return s
}
