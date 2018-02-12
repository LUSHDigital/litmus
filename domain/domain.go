/*Package internal contains code meant for internal consumption*/
package domain

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

// HttpStatusFmt returns a string formatted representation of an HTTP status code
// in a manner that is easily digestible by a human reader:
//  ex: 500 (Internal Server Error)
func HttpStatusFmt(code int) string {
	txt := http.StatusText(code)
	if txt == "" {
		txt = "INVALID RESPONSE CODE"
	}
	return fmt.Sprintf("%d (%s)", code, txt)
}

// Equals performs a deep equals against two objects and
// fails if they're not equal.
func Equals(tb testing.TB, exp interface{}, got interface{}) {
	if !reflect.DeepEqual(exp, got) {
		tb.Fatalf("\texp: %#v\n\tgot: %#v", exp, got)
	}
}
