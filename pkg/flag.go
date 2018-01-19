package pkg

import (
	"fmt"
	"strings"
)

// KeyValuePair allows a map to be expressed in slice form.
type KeyValuePair struct {
	Key   string
	Value string
}

// KeyValuePairs is a slice of KeyValuePair.
type KeyValuePairs []KeyValuePair

// Set adds an item to KeyValuePairs.
func (kvps *KeyValuePairs) Set(value string) (err error) {
	parts := strings.Split(value, "=")

	*kvps = append(*kvps, KeyValuePair{Key: parts[0], Value: parts[1]})
	return
}

// String returns the string representation of KeyValuePairs.
func (kvps *KeyValuePairs) String() string {
	return fmt.Sprintf("%v", *kvps)
}
