package test

import (
	"reflect"
	"testing"
)

// Equals performs a deep equal comparison against two
// values and fails if they are not the same.
func Equals(tb testing.TB, exp, act interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(exp, act) {
		tb.Fatalf("\n\texp: %#[1]v (%[1]T)\n\tgot: %#[2]v (%[2]T)\n", exp, act)
	}
}

// Assert checks that an expected outcome is satisfied
// and fails if it is not.
func Assert(tb testing.TB, exp bool) {
	tb.Helper()
	if !exp {
		tb.Fatal("\n\texpectation not met")
	}
}

// ErrorNil asserts that an error is nil and fails if it's not.
func ErrorNil(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("expected no error but got: %v", err)
	}
}
