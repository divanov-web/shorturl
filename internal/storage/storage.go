package storage

import (
	"math/rand"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
var urlStorage = make(map[string]string)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}
	return string(b)
}

func MakeShort(url string) string {
	id := generateID(8)
	urlStorage[id] = url
	return id
}

func GetURL(id string) (string, bool) {
	url, ok := urlStorage[id]
	return url, ok
}

// ForceSet используется для тестов
func ForceSet(id, url string) {
	urlStorage[id] = url
}
