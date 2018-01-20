package pkg

import (
	"net/http"
)

// HeaderGetter extracts information from response headers.
type HeaderGetter struct{}

// Get extracts a value out of request headers.
func (e *HeaderGetter) Get(c GetterConfig, header http.Header) (value string, err error) {
	for k, v := range header {
		if c.Path == k && len(v) > 0 {
			return v[0], nil
		}
	}
	return
}
