package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BuildCmd  string   `yaml:"build_cmd"`
	RunCmd    string   `yaml:"run_cmd"`
	AppPort   int      `yaml:"app_port"`
	ProxyPort int      `yaml:"proxy_port"`
	Watch     []string `yaml:"watch"`
	ExclDirs  []string `yaml:"excl_dirs"`
}

func DefaultConfig() Config {
	defaultConfig := Config{
		BuildCmd:  "go build -o tmp/main .",
		RunCmd:    "./tmp/main",
		AppPort:   3000,
		ProxyPort: 3001,
		Watch:     []string{".go", ".html", ".js", ".ts", ".css", ".tmpl"},
		ExclDirs:  []string{"node_modules", "tmp", "temp", ".git"},
	}

	return defaultConfig
}

func Load(filename string) (Config, error) {
	var config Config

	configFile, err := os.ReadFile(filename)

	if err != nil {
		return Config{}, err
	}

	unmarshalErr := yaml.Unmarshal(configFile, &config)

	if unmarshalErr != nil {
		return Config{}, unmarshalErr
	}

	return config, nil
}

func Save(filename string, cfg Config) error {
	file, err := yaml.Marshal(cfg)

	if err != nil {
		return err
	}

	fileErr := os.WriteFile(filename, file, 0644)

	if fileErr != nil {
		return fileErr
	}

	return nil
}
