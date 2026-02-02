package configs

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config хранит конфигурацию приложения
type Config struct {
	Port                  string
	CleanupInterval       time.Duration
	ArchiveAfter          time.Duration
	ReminderCheckInterval time.Duration
	LoggerBufferSize      int
}

// Load загружает конфигурацию из .env файла, переменных окружения и флагов
func Load() *Config {
	// Загрузить .env файл
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables only")
	}

	cfg := &Config{
		Port:                  getEnv("PORT", ""),
		CleanupInterval:       getDurationEnv("CLEANUP_INTERVAL", 0),
		ArchiveAfter:          getDurationEnv("ARCHIVE_AFTER", 0),
		ReminderCheckInterval: getDurationEnv("REMINDER_CHECK_INTERVAL", 0),
		LoggerBufferSize:      getIntEnv("LOGGER_BUFFER_SIZE", 0),
	}

	// Проверка обязательных параметров
	if cfg.Port == "" {
		log.Fatal("PORT is required. Set it in .env file or environment variable")
	}
	if cfg.CleanupInterval == 0 {
		log.Fatal("CLEANUP_INTERVAL is required. Set it in .env file or environment variable")
	}
	if cfg.ArchiveAfter == 0 {
		log.Fatal("ARCHIVE_AFTER is required. Set it in .env file or environment variable")
	}
	if cfg.ReminderCheckInterval == 0 {
		log.Fatal("REMINDER_CHECK_INTERVAL is required. Set it in .env file or environment variable")
	}
	if cfg.LoggerBufferSize == 0 {
		log.Fatal("LOGGER_BUFFER_SIZE is required. Set it in .env file or environment variable")
	}

	flag.StringVar(&cfg.Port, "port", cfg.Port, "Server port")
	flag.Parse()

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
