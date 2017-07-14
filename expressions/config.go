package expressions

// Config holds configuration information for expression interpretation.
type Config struct {
	filters map[string]interface{}
}

// NewConfig creates a new Config.
func NewConfig() Config {
	return Config{}
}
