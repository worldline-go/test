package testdb

import "strings"

func trim(s string) string {
	return strings.Fields(strings.TrimSpace(s))[0]
}
