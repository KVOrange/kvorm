package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Config struct {
	Server struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		FileRootPath string `yaml:"file_root_path"`
		FileRootUrl  string `yaml:"file_root_url"`
	} `yaml:"server"`
	Database struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		User      string `yaml:"user"`
		Password  string `yaml:"password"`
		Name      string `yaml:"name"`
		PollCount int32  `yaml:"poll_count"`
	} `yaml:"database"`
	Centrifugo struct {
		Endpoint string `yaml:"endpoint"`
		ApiKey   string `yaml:"api_key"`
	} `yaml:"centrifugo"`
	Security struct {
		PasswordSalt string `yaml:"password_salt"`
		JwtSecretKey string `yaml:"jwt_secret_key"`
	} `yaml:"security"`
}

func InitConfig() (*Config, error) {
	var path string
	ex, err := os.Executable()
	config := &Config{}

	configPath := os.Getenv("PATH_TO_CONFIG")
	exPath := filepath.Dir(ex)
	if configPath != "" {
		path = configPath
	} else {
		path = fmt.Sprintf("%s/%s", exPath, "config/config.yaml")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
