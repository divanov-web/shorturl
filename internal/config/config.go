package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() *Config {
	// флаги без дефолтов
	addrFlag := flag.String("a", "", "адрес запуска HTTP-сервера")
	baseFlag := flag.String("b", "", "базовый адрес для сокращённых URL")
	filePathFlag := flag.String("f", "", "путь к файлу хранения данных")

	flag.Parse()

	// переменные окружения
	envCfg := Config{}
	if err := env.Parse(&envCfg); err != nil {
		log.Fatal(err)
	}

	// Формируем итоговую конфигурацию с приоритетами:
	// 1. Переменная окружения
	// 2. Флаг командной строки
	// 3. Значение по умолчанию
	cfg := &Config{
		ServerAddress:   chooseValue(envCfg.ServerAddress, *addrFlag, "localhost:8080"),
		BaseURL:         chooseValue(envCfg.BaseURL, *baseFlag, "http://localhost:8080"),
		FileStoragePath: chooseValue(envCfg.FileStoragePath, *filePathFlag, "shortener_data.json"),
	}

	return cfg
}

// Вспомогательная функция для выбора значения
func chooseValue(envVal, flagVal, defaultVal string) string {
	if envVal != "" {
		return envVal
	}
	if flagVal != "" {
		return flagVal
	}
	return defaultVal
}
