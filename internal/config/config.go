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
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func NewConfig() *Config {
	// Загрузим .env только если переменные ещё не заданы в окружении
	//_ = godotenv.Load()

	// флаги без дефолтов
	addrFlag := flag.String("a", "", "адрес запуска HTTP-сервера")
	baseFlag := flag.String("b", "", "базовый адрес для сокращённых URL")
	filePathFlag := flag.String("f", "", "путь к файлу хранения данных")
	dbDSNFlag := flag.String("d", "", "строка подключения к БД")

	flag.Parse()

	envCfg := Config{}
	if err := env.Parse(&envCfg); err != nil {
		log.Fatal(err)
	}

	cfg := &Config{
		ServerAddress:   chooseValue(envCfg.ServerAddress, *addrFlag, "localhost:8080"),
		BaseURL:         chooseValue(envCfg.BaseURL, *baseFlag, "http://localhost:8080"),
		FileStoragePath: chooseValue(envCfg.FileStoragePath, *filePathFlag, "shortener_data.json"),
		DatabaseDSN:     chooseValue(envCfg.DatabaseDSN, *dbDSNFlag, ""),
	}

	return cfg
}

func chooseValue(envVal, flagVal, defaultVal string) string {
	if envVal != "" {
		return envVal
	}
	if flagVal != "" {
		return flagVal
	}
	return defaultVal
}
