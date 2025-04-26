package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func NewConfig() *Config {
	cfg := &Config{
		ServerAddress: "default.com:8080",
		BaseURL:       "https://default.com:8080",
	}

	// Загружаем переменные окружения
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("ошибка чтения переменных окружения: %v", err)
	}

	envAddr := cfg.ServerAddress
	envBase := cfg.BaseURL

	// Переменные из флагов
	addrFlag := flag.String("a", cfg.ServerAddress, "адрес запуска HTTP-сервера")
	baseFlag := flag.String("b", cfg.BaseURL, "базовый адрес для сокращённых URL")

	flag.Parse()

	// Если значение из env пустое — берём флаг
	if envAddr == "" && addrFlag != nil {
		cfg.ServerAddress = *addrFlag
	}
	if envBase == "" && baseFlag != nil {
		cfg.BaseURL = *baseFlag
	}

	return cfg
}
