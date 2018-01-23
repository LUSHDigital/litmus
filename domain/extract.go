// Package extract defines all the possible extraction methods that can be applied on a response.
//
// Signatures
//
// it is important for consistency that each extractor follow the same function signature:
//  func(r RequestTest, resp *http.Response, env map[string]interface{})
//
//
// this will allow easier refactoring or interfacing later on if this becomes necessary.
package domain

import (
	"io/ioutil"
	"net/http"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

var yellow = color.New(color.FgHiYellow).SprintFunc()

func ProcessResponse(r *RequestTest, resp *http.Response, env map[string]interface{}) error {
	if err := StatusCode(r, resp, env); err != nil {
		return err
	}
	if err := Header(r, resp, env); err != nil {
		return err
	}

	return Body(r, resp, env)
}

// StatusCode - extracts the status code and checks it against the expected value
func StatusCode(r *RequestTest, resp *http.Response, _ map[string]interface{}) error {
	if resp == nil {
		return errors.New("unexpected nil response")
	}
	if r.WantsCode != 0 && r.WantsCode != resp.StatusCode {
		return errors.Errorf("expected response code: %s, but got: %s",
			HttpStatusFmt(r.WantsCode),
			HttpStatusFmt(resp.StatusCode),
		)
	}
	return nil
}

// Payload - checks the body against the expected value
func Body(r *RequestTest, resp *http.Response, env map[string]interface{}) error {
	if resp == nil {
		return errors.New("unexpected nil response")
	}

	// If we're unable to ascertain the body type, we won't
	// be able to extract anything and needn't bother reading
	// the response body.
	bodyGetter, err := NewBodyGetter(resp)
	if err != nil {
		return errors.Wrap(err, "creating body getter")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	defer resp.Body.Close()

	for k, v := range r.Body {
		path, expected, set, err := extractParam(k, v)
		if err != nil {
			return errors.Wrap(err, "extracting body param")
		}

		actual, err := bodyGetter.Get(path, respBody)
		if err != nil {
			return err
		}

		if err = equals(expected, actual); err != nil {
			return errors.Wrap(err, "assertion failed")
		}

		if set != "" {
			env[set] = actual
		}
	}

	return nil
}

func getFirst(m map[string]interface{}) (key, val string, err error) {
	for k, v := range m {
		val, ok := v.(string)
		if !ok {
			return "", "", errors.Errorf("expected string but got: %T", val)
		}
		return k, val, nil
	}
	return "", "", nil
}

func extractParam(key string, value interface{}) (path, expected, set string, err error) {
	switch x := value.(type) {
	case map[string]interface{}:
		path, expected, err = getFirst(x)
		if err != nil {
			return
		}
		return path, expected, key, err
	case string:
		return key, x, "", err
	default:
		return "", "", "", errors.Errorf("expected string but got: %T", x)
	}
}

// Header - extracts a header value and checks it against the expected value
func Header(r *RequestTest, resp *http.Response, env map[string]interface{}) error {
	if resp == nil {
		return errors.New("unexpected nil response")
	}

	headerGetter := &HeaderGetter{}

	for k, v := range r.Head {
		path, expected, set, err := extractParam(k, v)
		if err != nil {
			return errors.Wrap(err, "extracting body param")
		}
		actual, err := headerGetter.Get(path, resp.Header)
		if err != nil {
			return err
		}

		if err = equals(expected, actual); err != nil {
			return errors.Wrap(err, "assertion failed")
		}

		if set != "" {
			env[set] = actual
		}
	}
	return nil
}

func equals(exp string, act string) (err error) {
	if exp != act {
		return errors.Errorf("\n\texp: %v\n\tgot: %v", exp, act)
	}
	return
}
