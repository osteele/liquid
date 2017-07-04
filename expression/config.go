package expression

// Config holds configuration information for expression interpretation.
type Config struct {
	filters *filterDictionary
}

// NewConfig creates a new Settings.
func NewConfig() Config {
	return Config{newFilterDictionary()}
}

// AddFilter adds a filter function to settings.
func (s Config) AddFilter(name string, fn interface{}) {
	s.filters.addFilter(name, fn)
}
