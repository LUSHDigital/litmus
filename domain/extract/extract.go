// Package extract defines all the possible extraction methods that can be applied on a response.
//
// Signatures
//
// it is important for consistency that each extractor follow the same function signature:
//  func(r format.RequestTest, resp *http.Response, env map[string]interface{})
//
//
// this will allow easier refactoring or interfacing later on if this becomes necessary.
package extract

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/LUSHDigital/litmus/domain"
	"github.com/LUSHDigital/litmus/format"
	"github.com/LUSHDigital/litmus/p"
	"github.com/LUSHDigital/litmus/pkg"
	"github.com/pkg/errors"
)

func ProcessResponse(r *format.RequestTest, resp *http.Response, env map[string]interface{}) error {
	if err := StatusCode(r, resp, env); err != nil {
		return err
	}
	if err := Header(r, resp, env); err != nil {
		return err
	}

	return Body(r, resp, env)
}

// StatusCode - extracts the status code and checks it against the expected value
func StatusCode(r *format.RequestTest, resp *http.Response, _ map[string]interface{}) error {
	if resp == nil {
		return errors.New("unexpected nil response")
	}
	if r.WantsCode != 0 && r.WantsCode != resp.StatusCode {
		return errors.Errorf("expected response code: %s, but got: %s",
			domain.HttpStatusFmt(r.WantsCode),
			domain.HttpStatusFmt(resp.StatusCode),
		)
	}
	return nil
}

// Body - checks the body against the expected value
func Body(r *format.RequestTest, resp *http.Response, env map[string]interface{}) error {
	if resp == nil {
		return errors.New("unexpected nil response")
	}
	getters := r.Getters.Filter("body")
	if len(getters) == 0 {
		return nil
	}

	// If we're unable to ascertain the body type, we won't
	// be able to extract anything and needn't bother reading
	// the response body.
	bodyGetter, err := pkg.NewBodyGetter(resp)
	if err != nil {
		return errors.Wrap(err, "creating body getter")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	defer resp.Body.Close()

	for _, getter := range getters {
		act, err := bodyGetter.Get(getter, respBody)
		if err != nil {
			return err
		}

		if getter.Expected != "" {
			if err = equals(getter.Expected, act); err != nil {
				return errors.Wrap(err, "assertion failed")
			}
		}

		if getter.Set != "" {
			if env == nil {
				return errors.Errorf("error setting environment variable %s", getter.Set)
			}
			env[getter.Set] = act
			fmt.Printf("\t[%s]  %s -> %s\n", p.Yellow("SET"), act, getter.Set)
		}
	}

	return nil
}

// Header - extracts a header value and checks it against the expected value
func Header(r *format.RequestTest, resp *http.Response, env map[string]interface{}) error {
	if resp == nil {
		return errors.New("unexpected nil response")
	}
	getters := r.Getters.Filter("head")
	if len(getters) == 0 {
		return nil
	}

	headerGetter := &pkg.HeaderGetter{}

	for _, getter := range getters {
		act, err := headerGetter.Get(getter, resp.Header)
		if err != nil {
			return err
		}

		if getter.Expected != "" {
			if err = equals(getter.Expected, act); err != nil {
				return errors.Wrap(err, "assertion failed")
			}
		}

		if getter.Set != "" {
			if env == nil {
				return errors.Errorf("error setting environment variable %s", getter.Set)
			}
			env[getter.Set] = act
			fmt.Printf("\t[%s]  %s -> %s\n", p.Yellow("SET"), act, getter.Set)
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
