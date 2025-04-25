package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	type want struct {
		address string
		baseURL string
	}

	tests := []struct {
		name string
		args []string
		want want
	}{
		{
			name: "custom flags",
			args: []string{"cmd", "-a", "localhost:9999", "-b", "http://test.url/"},
			want: want{
				address: "localhost:9999",
				baseURL: "http://test.url/",
			},
		},
		{
			name: "default flags",
			args: []string{"cmd"}, // без параметров
			want: want{
				address: "localhost:8080",
				baseURL: "http://localhost:8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// сохраняем оригинальные аргументы, чтобы не ломать другие тесты
			origArgs := os.Args
			defer func() { os.Args = origArgs }()

			// подставляем тестовые аргументы
			os.Args = tt.args

			// сбрасываем глобальные флаги
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg := NewConfig()

			assert.Equal(t, tt.want.address, cfg.ServerAddress)
			assert.Equal(t, tt.want.baseURL, cfg.BaseURL)
		})
	}
}
