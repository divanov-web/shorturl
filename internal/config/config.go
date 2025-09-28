package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config структура с главным конфигом приложения
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	AuthSecret      string `env:"AUTH_SECRET" json:"auth_secret"`
	StorageType     string //определяется автоматически
	PprofMode       bool   `env:"PPROF_MODE" json:"pprof_mode"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigPath      string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"` // CIDR доверенной подсети
}

// NewConfig Создаёт конфиг приложения и возвращает в виде структуры
func NewConfig() *Config {
	// Загрузим .env только если переменные ещё не заданы в окружении
	_ = godotenv.Load()

	// флаги без дефолтов
	addrFlag := flag.String("a", "", "адрес запуска HTTP-сервера")
	baseFlag := flag.String("b", "", "базовый адрес для сокращённых URL")
	filePathFlag := flag.String("f", "", "путь к файлу хранения данных")
	dbDSNFlag := flag.String("d", "", "строка подключения к БД")
	authSecretFlag := flag.String("auth-secret", "", "секрет для подписи JWT")
	pprofFlag := flag.Bool("pprof", false, "включить pprof-сервер")
	httpsFlag := flag.Bool("s", false, "включить HTTPS-сервер")
	cfgPathFlag := flag.String("c", "", "путь к JSON-файлу конфигурации")
	flag.StringVar(cfgPathFlag, "config", "", "путь к JSON-файлу конфигурации")
	trustedSubnetFlag := flag.String("t", "", "CIDR доверенной подсети (например, 192.168.0.0/24)")

	flag.Parse()

	//Конфиг из переменных окружения
	envCfg := Config{}
	if err := env.Parse(&envCfg); err != nil {

		log.Fatal(err)
	}

	//Конфиг из файла (самый низкий приоритет)
	cfgFromFile := Config{}
	configPath := chooseValue(envCfg.ConfigPath, *cfgPathFlag, "", "")
	if configPath != "" {
		if f, err := os.Open(configPath); err == nil {
			defer f.Close()
			if err := json.NewDecoder(f).Decode(&cfgFromFile); err != nil {
				log.Printf("config: can't decode JSON config %q: %v", configPath, err)
			}
		} else {
			log.Printf("config: can't open config file %q: %v", configPath, err)
		}
	}

	cfg := &Config{
		ServerAddress:   chooseValue(envCfg.ServerAddress, *addrFlag, cfgFromFile.ServerAddress, "localhost:8080"),
		BaseURL:         chooseValue(envCfg.BaseURL, *baseFlag, cfgFromFile.BaseURL, "http://localhost:8080"),
		FileStoragePath: chooseValue(envCfg.FileStoragePath, *filePathFlag, cfgFromFile.FileStoragePath, "shortener_data.json"),
		DatabaseDSN:     chooseValue(envCfg.DatabaseDSN, *dbDSNFlag, cfgFromFile.DatabaseDSN, ""),
		AuthSecret:      chooseValue(envCfg.AuthSecret, *authSecretFlag, cfgFromFile.AuthSecret, "dev-secret-key"),
		PprofMode:       envCfg.PprofMode || *pprofFlag || cfgFromFile.PprofMode,
		EnableHTTPS:     envCfg.EnableHTTPS || *httpsFlag || cfgFromFile.EnableHTTPS,
		TrustedSubnet:   chooseValue(envCfg.TrustedSubnet, *trustedSubnetFlag, cfgFromFile.TrustedSubnet, ""),
	}

	// при HTTPS меняем протокол
	if cfg.EnableHTTPS && strings.HasPrefix(cfg.BaseURL, "http://") {
		cfg.BaseURL = "https://" + strings.TrimPrefix(cfg.BaseURL, "http://")
	}

	cfg.StorageType = detectStorageType(cfg.DatabaseDSN, cfg.FileStoragePath)

	return cfg
}

// chooseValue определяет очерёдность параметров конфига
func chooseValue(envVal, flagVal, fileVal, defaultVal string) string {
	if envVal != "" {
		return envVal
	}
	if flagVal != "" {
		return flagVal
	}
	if fileVal != "" {
		return fileVal
	}
	return defaultVal
}

func detectStorageType(dsn, filePath string) string {
	if dsn != "" {
		return "postgres"
	}
	if filePath != "" {
		return "file"
	}
	return "memory"
}
