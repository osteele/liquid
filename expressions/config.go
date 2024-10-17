package expressions

// Config holds configuration information for expression interpretation.
type Config struct {
	filters map[string]any
}

// NewConfig creates a new Config.
func NewConfig() Config {
	return Config{}
}
