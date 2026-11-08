package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	Port      string
	JWTSecret []byte
	JWTTTL    time.Duration
	RefreshTTL time.Duration
}

func Load() Config {
	port := os.Getenv("APP_PORT")
	if port == "" { port = "8080" }

	secret := os.Getenv("JWT_SECRET")
	if secret == "" { log.Fatal("JWT_SECRET is required") }

	ttl := os.Getenv("JWT_TTL")
	if ttl == "" { ttl = "15m" } // По умолчанию 15 минут для Access
	dur, err := time.ParseDuration(ttl)
	if err != nil { log.Fatal("bad JWT_TTL") }

	refreshTTL := os.Getenv("REFRESH_TTL")
	if refreshTTL == "" { refreshTTL = "168h" } // 7 дней
	rDur, err := time.ParseDuration(refreshTTL)
	if err != nil { log.Fatal("bad REFRESH_TTL") }

	return Config{Port: ":" + port, JWTSecret: []byte(secret), JWTTTL: dur, RefreshTTL: rDur}
}