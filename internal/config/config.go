package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DB_config_path string
	HTTPServer     `yaml:"http_server"`
}

func SetConfig() (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Loading Config failed, error: %w", err.Error())
		return "", err
	}
	DB_config_path := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	return DB_config_path, nil
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"0.0.0.0:8082"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoadConfig(server_configPath string) (Config, error) {
	db_configPath, err := SetConfig()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(server_configPath)
	if err != nil {
		log.Fatal("Reading Config file failed", map[string]string{
			"error": err.Error(),
		})
		return Config{}, fmt.Errorf("не удалось прочитать файл %s: %w", server_configPath, err)
	}

	var port Config
	err = yaml.Unmarshal(data, &port)
	if err != nil {
		log.Fatal("Parsing YAML failed", map[string]string{
			"error": err.Error(),
		})
		return Config{}, fmt.Errorf("не удалось распарсить YAML: %w", err)
	}

	return Config{
		DB_config_path: db_configPath,
		HTTPServer: HTTPServer{
			Address:     port.Address,
			Timeout:     port.Timeout,
			IdleTimeout: port.IdleTimeout,
		},
	}, nil
}
