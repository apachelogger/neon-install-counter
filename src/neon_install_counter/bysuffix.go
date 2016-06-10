package neon_install_counter

// gets integer from log filename
import (
	"strconv"
	"strings"
)

type BySuffix []string

func (s BySuffix) Len() int {
	return len(s)
}
func (s BySuffix) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s BySuffix) Less(i, j int) bool {
	return s.toInteger(s[i]) < s.toInteger(s[j])
}
func (s BySuffix) toInteger(str string) int {
	parts := strings.Split(str, ".")
	index := 1
	if parts[len(parts)-index] == "gz" {
		index = 2
	}
	part := parts[len(parts)-index]
	a, err := strconv.Atoi(part)
	if err != nil {
		return -1
	}
	return a
}
