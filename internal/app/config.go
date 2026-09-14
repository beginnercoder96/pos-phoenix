package app

import "os"

type Config struct {
	Addr, DatabasePath, AdminEmail, AdminPassword, BusinessTimezone string
	SecureCookies                                                   bool
}

func ConfigFromEnv() Config {
	return Config{
		Addr:             getenv("ADDR", ":8080"),
		DatabasePath:     getenv("DATABASE_PATH", "pos.db"),
		AdminEmail:       os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		AdminPassword:    os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		BusinessTimezone: getenv("BUSINESS_TIMEZONE", "Asia/Jakarta"),
		SecureCookies:    os.Getenv("SESSION_SECURE") == "true",
	}
}
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
