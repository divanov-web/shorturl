package config

import (
	"flag"
)

type Config struct {
	ServerAddress string // адрес HTTP-сервера (флаг -a)
	BaseURL       string // базовый адрес для сокращённых ссылок (флаг -b)
}

func NewConfig() *Config {
	addr := flag.String("a", "localhost:8080", "адрес запуска HTTP-сервера")
	base := flag.String("b", "http://localhost:8080", "базовый адрес для сокращённых URL")

	flag.Parse()

	return &Config{
		ServerAddress: *addr,
		BaseURL:       *base,
	}
}
