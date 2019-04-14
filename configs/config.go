package configs

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Config struct
type Config struct {
	Debug    bool     `yaml:"debug"`
	Interval uint16   `yaml:"interval"`
	LogPath  string   `yaml:"log_path"`
	OutPath  string   `yaml:"out_path"`
	Rooms    []string `yaml:"rooms"`
}

// Parse Configuration file
func Parse(path string) (*Config, error) {
	// Read Configuration file
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read configuration file - %s", path)
	}

	c := &Config{}
	// Parse yaml
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		return nil, fmt.Errorf("Configuration file parsing failed - %s", path)
	}

	return c, nil
}
