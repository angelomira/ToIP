package config

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = 12
const maxRecommendedBcryptCost = 14

// Config contains the runtime settings read from environment variables.
type Config struct {
	DBDSN      string
	BcryptCost int
	Addr       string
}

func Load() (Config, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}

	cost := defaultBcryptCost
	if value := os.Getenv("BCRYPT_COST"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("BCRYPT_COST must be an integer: %w", err)
		}
		cost = parsed
	}
	if cost < bcrypt.MinCost || cost > maxRecommendedBcryptCost {
		return Config{}, fmt.Errorf(
			"BCRYPT_COST must be between %d and %d for this project",
			bcrypt.MinCost,
			maxRecommendedBcryptCost,
		)
	}

	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{
		DBDSN:      dsn,
		BcryptCost: cost,
		Addr:       addr,
	}, nil
}
