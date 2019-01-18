package helm

import "strings"

// Transform performs a string replacement of the specified source for
// a given key with the replacement string
func Transform(src string, key string, replacement string) []byte {
	return []byte(strings.Replace(src, key, replacement, -1))
}
