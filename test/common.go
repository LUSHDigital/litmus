package test

import (
	"reflect"
	"testing"
)

// Equals performs a deep equals against two objects and
// fails if they're not equal.
func Equals(tb testing.TB, exp interface{}, got interface{}) {
	if !reflect.DeepEqual(exp, got) {
		tb.Fatalf("\texp: %#v\n\tgot: %#v", exp, got)
	}
}
