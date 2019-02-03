package config

// Config object
type Config struct {
	Host string
	Port int
}

var defaultConf = Config{
	Host: "0.0.0.0",
	Port: 4444,
}

var globalConf = defaultConf

// NewConfig create default config
func NewConfig() *Config {
	conf := defaultConf
	return &conf
}

// GetGlobalConfig returns global config
func GetGlobalConfig() *Config {
	return &globalConf
}