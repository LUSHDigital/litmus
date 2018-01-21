// Package format defines the format of a Litmus test file, as represented in code.
package format

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
	Name      string            `toml:"name"`
	Method    string            `toml:"method"`
	URL       string            `toml:"url"`
	Headers   map[string]string `toml:"headers"`
	Body      string            `toml:"body"`
	Getters   GetterConfigs     `toml:"getters"`
	WantsCode int               `toml:"wants_code"`
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
