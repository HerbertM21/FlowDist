package db

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var loadEnvOnce sync.Once

func LoadEnvFile(path string) {
	loadEnvOnce.Do(func() {
		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, `"`)

			if key == "" {
				continue
			}

			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
	})
}

func envIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}

	return n
}

func ResolveDSN(dsn string) string {
	LoadEnvFile(".env")

	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if host == "" {
			host = "192.168.1.12"
		}
		if port == "" {
			port = "5433"
		}
		if user == "" {
			user = "postgres"
		}
		if password == "" {
			password = "postgres"
		}
		if name == "" {
			name = "distribuidora"
		}
		if sslmode == "" {
			sslmode = "disable"
		}

		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + name + " sslmode=" + sslmode
	}

	return dsn
}

func InitDB(dsn string) (*sqlx.DB, error) {
	dsn = ResolveDSN(dsn)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	maxOpenConns := envIntOrDefault("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := envIntOrDefault("DB_MAX_IDLE_CONNS", 25)

	// Optimize connection pool for local LAN usage (low RAM footprint)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("Connected to PostgreSQL (Pool: max_open=%d max_idle=%d)", maxOpenConns, maxIdleConns)
	return db, nil
}
