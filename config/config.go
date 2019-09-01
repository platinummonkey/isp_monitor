package config

// Section defines the config section.
type Section struct {
	Name     string                 `yaml:"name"`
	Type     string                 `yaml:"type"`
	Interval string                 `yaml:"interval"`
	Options  map[string]interface{} `yaml:"options"`
}

// Config defines the configuration
type Config struct {
	Collectors []Section `yaml:"collectors,inline"`
	Reporters  []Section `yaml:"reporters,inline"`
}
