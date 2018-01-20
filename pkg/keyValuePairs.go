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

// KeyValuePairs is the slice type for KeyValuePair
type KeyValuePairs []KeyValuePair

// String returns the string representation of KeyValuePairs.
func (k *KeyValuePairs) String() string {
	return fmt.Sprintf("%#+v", k)
}

// Set adds an item to KeyValuePairs.
func (k *KeyValuePairs) Set(value string) error {
	parts := strings.Split(value, "=")

	*k = append(*k, KeyValuePair{Key: parts[0], Value: parts[1]})
	return nil
}

// Type returns a string representation of the type of KeyValuePairs
func (k *KeyValuePairs) Type() string {
	return fmt.Sprintf("%T", k)
}
