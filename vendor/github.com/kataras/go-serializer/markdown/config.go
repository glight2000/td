package markdown

import (
	"github.com/imdario/mergo"
)

// Config is the configuration for this serializer
type Config struct {
	MarkdownSanitize bool
}

// DefaultConfig returns the default configuration for this serializer
func DefaultConfig() Config {
	return Config{
		MarkdownSanitize: false,
	}
}

// Merge merges the default with the given config and returns the result
func (c Config) Merge(cfg []Config) (config Config) {

	if len(cfg) > 0 {
		config = cfg[0]
		mergo.Merge(&config, c)
	} else {
		_default := c
		config = _default
	}

	return
}

// MergeSingle merges the default with the given config and returns the result
func (c Config) MergeSingle(cfg Config) (config Config) {

	config = cfg
	mergo.Merge(&config, c)

	return
}
