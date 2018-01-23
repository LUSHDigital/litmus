// Package format defines the format of a Litmus test file, as represented in code.
package domain

import (
	"bytes"
	"github.com/tidwall/sjson"
	"text/template"
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
	Body          string                 `toml:"body"`
	BodyModifiers map[string]interface{} `toml:"bodymod"`
	Getters       GetterConfigs          `toml:"getters"`
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

// Filter filters a slice of getter configs by a
// given type.
func (c GetterConfigs) Filter(t string) (filtered GetterConfigs) {
	for _, config := range c {
		if config.Type == t {
			filtered = append(filtered, config)
		}
	}
	return
}

func (r *RequestTest) ApplyEnv(env map[string]interface{}) (err error) {
	if r.URL, err = applyTpl(r.URL, env); err != nil {
		return
	}
	if r.Body, err = applyTpl(r.Body, env); err != nil {
		return
	}
	for k, v := range r.BodyModifiers {
		if r.Body, err = sjson.Set(r.Body, k, v); err != nil {
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
	for i := range r.Getters {
		if r.Getters[i].Expected, err = applyTpl(r.Getters[i].Expected, env); err != nil {
			return
		}
		if r.Getters[i].Path, err = applyTpl(r.Getters[i].Path, env); err != nil {
			return
		}
	}
	return
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
