package domain

import (
	"net/http"

	"github.com/pkg/errors"
	"strings"
)

// HeaderGetter extracts information from response headers.
type HeaderGetter struct{}

// Get extracts a value out of request headers.
func (e *HeaderGetter) Get(path string, header http.Header) (value string, err error) {
	for k, v := range header {
		if strings.ToLower(path) == strings.ToLower(k) && len(v) > 0 {
			return v[0], nil
		}
	}
	return "", errors.New("no matching header found")
}
