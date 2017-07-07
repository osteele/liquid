package expression

// Config holds configuration information for expression interpretation.
type Config struct {
	filters map[string]interface{}
}

// NewConfig creates a new Settings.
func NewConfig() Config {
	return Config{}
}
