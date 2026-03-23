package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Save marshals the Config to YAML and writes it to the given path.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
