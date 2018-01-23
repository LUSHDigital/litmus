package domain

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// BodyGetter defines the behavior or something that can
// extract information from a response body.
type BodyGetter interface {
	Get(c GetterConfig, body []byte) (value string, err error)
}

// NewBodyGetter returns the body extracter based on the
// Content-Type found in the response headers.
func NewBodyGetter(resp *http.Response) (e BodyGetter, err error) {
	contentType := resp.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return &JSONBodyGetter{}, nil
	default:
		return nil, errors.Errorf("invalid Content-Type %q", contentType)
	}
}

// JSONBodyGetter extracts information from a response
// body using JSON dot notation.
type JSONBodyGetter struct{}

// Get extracts a value out of a JSON body using JSON
// dot notation.
func (e *JSONBodyGetter) Get(c GetterConfig, body []byte) (value string, err error) {
	result := gjson.GetBytes(body, c.Path)
	if !result.Exists() {
		return "", errors.Errorf("no value at path %q in JSON body", c.Path)
	}

	return result.String(), nil
}
