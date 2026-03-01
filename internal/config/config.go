package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {

	DatabaseURL string

	RedisAddr string

	WorkerMin int
	WorkerMax int

	TokenRate  int
	TokenBurst int

	APIRateLimit int

	Port string
}

func LoadConfig() Config {

	err := godotenv.Load()

	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return Config{

		DatabaseURL: getEnv("DATABASE_URL",""),

		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),

		WorkerMin: getEnvInt("WORKER_MIN", 1),
		WorkerMax: getEnvInt("WORKER_MAX", 5),

		TokenRate: getEnvInt("TOKEN_RATE", 3),
		TokenBurst: getEnvInt("TOKEN_BURST", 5),

		APIRateLimit: getEnvInt("API_RATE_LIMIT", 30),

		Port: getEnv("PORT", "8080"),
	}

}

func getEnv(key string, defaultVal string) string {

	val := os.Getenv(key)

	if val == "" {
		return defaultVal
	}

	return val
}

func getEnvInt(key string, defaultVal int) int {

	valStr := os.Getenv(key)

	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(valStr)

	if err != nil {
		return defaultVal
	}

	return val
}