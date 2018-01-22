/*Package internal contains code meant for internal consumption*/
package domain

import (
	"fmt"
	"net/http"
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
