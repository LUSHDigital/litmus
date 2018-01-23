// Package format defines the format of a Litmus test file, as represented in code.
package domain

import (
	"bytes"
	"github.com/tidwall/sjson"
	"text/template"
	"fmt"
)

// TestFile is the top level container element defining a Litmus test file
type TestFile struct {
	// Litmus is the top level table
	Litmus struct {
		// test is singular to enable singular dot notation in the file
		Test []RequestTest
	}
}

// RequestTest defines all the necessary fields to define a Litmus test
type RequestTest struct {
	Name          string                 `toml:"name"`
	Method        string                 `toml:"method"`
	URL           string                 `toml:"url"`
	Headers       map[string]string      `toml:"headers"`
	Query         map[string]string      `toml:"query"`
	Payload       string                 `toml:"payload"`
	BodyModifiers map[string]interface{} `toml:"bodymod"`
	Body          map[string]interface{} `toml:"body"`
	Head          map[string]interface{} `toml:"head"`
	WantsCode     int                    `toml:"wants_code"`
}

// GetterConfigs is a slice of GetterConfig
type GetterConfigs []GetterConfig

// GetterConfig provides the information required
// to get data from a response.
type GetterConfig struct {
	Path     string `toml:"path"`
	Set      string `toml:"set"`
	Type     string `toml:"type"`
	Expected string `toml:"exp"`
}

func (r *RequestTest) ApplyEnv(env map[string]interface{}) (err error) {
	if r.URL, err = applyTpl(r.URL, env); err != nil {
		return
	}
	if r.Payload, err = applyTpl(r.Payload, env); err != nil {
		return
	}
	for k, v := range r.BodyModifiers {
		if r.Payload, err = sjson.Set(r.Payload, k, v); err != nil {
			return
		}
	}
	for k, v := range r.Headers {
		if r.Headers[k], err = applyTpl(v, env); err != nil {
			return
		}
	}
	for k, v := range r.Query {
		if r.Query[k], err = applyTpl(v, env); err != nil {
			return
		}
	}

	if err := modifyRequestEnv(r.Body, env); err != nil {
		return err
	}
	if err := modifyRequestEnv(r.Head, env); err != nil {
		return err
	}
	return
}

func modifyRequestEnv(requestEnv map[string]interface{}, globalEnv map[string]interface{}) error {
	for k, v := range requestEnv {
		// handle maps
		if x, ok := v.(map[string]interface{}); ok {
			if err := handleMap(x, globalEnv); err != nil {
				return err
			}
			continue
		}

		// otherwise continue as normal
		key, err := applyTpl(k, globalEnv)
		if err != nil {
			//return err
		}

		val := fmt.Sprintf("%v", v)
		value, err := applyTpl(val, globalEnv)
		if err != nil {
			//return err
		}

		requestEnv[key] = value
		if key != k {
			delete(requestEnv, k)
		}
	}
	return nil
}

func handleMap(m, globalEnv map[string]interface{}) error {
	for k, v := range m {
		key, err := applyTpl(k, globalEnv)
		if err != nil {
			return err
		}

		val, ok := v.(string)
		if !ok {
			fmt.Errorf("expected string but got %T", val)
		}

		result, err := applyTpl(val, globalEnv)
		if err != nil {
			return err
		}
		m[key] = result
		if key != k {
			delete(m, k)
		}
	}
	return nil
}

func applyTpl(input string, env map[string]interface{}) (output string, err error) {
	buf := &bytes.Buffer{}
	t, err := template.New("anon").Parse(input)
	if err != nil {
		return "", err
	}
	if err = t.Execute(buf, env); err != nil {
		return
	}

	return buf.String(), nil
}
