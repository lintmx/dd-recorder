package configs

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Config struct
type Config struct {
	Debug    bool     `yaml:"debug"`
	Interval uint16   `yaml:"interval"`
	LogPath  string   `yaml:"log_path"`
	OutPath  string   `yaml:"out_path"`
	Rooms    []string `yaml:"rooms"`
}

// InitConfig return a config with parse
func InitConfig(conf string) *Config {
	config := &Config{}

	file, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read configuration file - %s\n", conf)
		os.Exit(1)
	}

	// Parse yaml
	err = yaml.Unmarshal(file, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration file parsing failed - %s\n", conf)
		os.Exit(1)
	}

	return config
}
