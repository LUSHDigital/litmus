package pkg

// RequestConfig describes a request to make.
type RequestConfig struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
	Getters GetterConfigs     `yaml:"getters"`
}

// GetterConfig provides the information required
// to get data from a response.
type GetterConfig struct {
	Path     string `yaml:"path"`
	Set      string `yaml:"set"`
	Type     string `yaml:"type"`
	Expected string `yaml:"exp"`
}

// GetterConfigs is a slice of GetterConfig.
type GetterConfigs []GetterConfig

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
