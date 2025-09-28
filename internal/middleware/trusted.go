package middleware

import (
	"net"
	"net/http"
	"strings"
)

// WithTrustedSubnet ограничивает доступ к хендлеру только клиентам,
// чей IP (из заголовка X-Real-IP) входит в заданный CIDR.
// При пустом cidr доступ запрещён всегда.
func WithTrustedSubnet(cidr string) func(http.Handler) http.Handler {
	var (
		_, ipnet, parseErr = net.ParseCIDR(strings.TrimSpace(cidr))
		cidrProvided       = strings.TrimSpace(cidr) != ""
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если подсеть не задана — запретить доступ
			if !cidrProvided || parseErr != nil || ipnet == nil {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			ipStr := strings.TrimSpace(r.Header.Get("X-Real-IP"))
			if ipStr == "" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil || !ipnet.Contains(ip) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
