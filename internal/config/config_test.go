package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("load default configuration", func(t *testing.T) {
		// Clear environment variables
		clearEnvVars()

		cfg := Load()

		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "postgres", cfg.Database.User)
		assert.Equal(t, "", cfg.Database.Password)
		assert.Equal(t, "libmngmt", cfg.Database.Name)
		assert.Equal(t, "disable", cfg.Database.SSLMode)
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("load configuration from environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("DB_HOST", "custom-db-host")
		os.Setenv("DB_PORT", "3306")
		os.Setenv("DB_USER", "custom-user")
		os.Setenv("DB_PASSWORD", "custom-pass")
		os.Setenv("DB_NAME", "custom-db")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("SERVER_HOST", "127.0.0.1")
		os.Setenv("SERVER_PORT", "9090")
		os.Setenv("LOG_LEVEL", "debug")

		cfg := Load()

		assert.Equal(t, "custom-db-host", cfg.Database.Host)
		assert.Equal(t, 3306, cfg.Database.Port)
		assert.Equal(t, "custom-user", cfg.Database.User)
		assert.Equal(t, "custom-pass", cfg.Database.Password)
		assert.Equal(t, "custom-db", cfg.Database.Name)
		assert.Equal(t, "require", cfg.Database.SSLMode)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 9090, cfg.Server.Port)
		assert.Equal(t, "debug", cfg.LogLevel)

		// Clean up
		clearEnvVars()
	})

	t.Run("load with partial environment variables", func(t *testing.T) {
		// Clear environment variables first
		clearEnvVars()

		// Set only some environment variables
		os.Setenv("DB_HOST", "partial-host")
		os.Setenv("SERVER_PORT", "7777")

		cfg := Load()

		// Should use env var where set
		assert.Equal(t, "partial-host", cfg.Database.Host)
		assert.Equal(t, 7777, cfg.Server.Port)

		// Should use defaults for unset vars
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "postgres", cfg.Database.User)
		assert.Equal(t, "", cfg.Database.Password)
		assert.Equal(t, "libmngmt", cfg.Database.Name)
		assert.Equal(t, "disable", cfg.Database.SSLMode)
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, "info", cfg.LogLevel)

		// Clean up
		clearEnvVars()
	})

	t.Run("configuration structure", func(t *testing.T) {
		cfg := Load()

		// Test that all fields are accessible
		assert.NotNil(t, cfg.Database)
		assert.NotNil(t, cfg.Server)
		assert.NotEmpty(t, cfg.LogLevel)
	})
}

func TestDatabaseConfig_Structure(t *testing.T) {
	t.Run("database config fields", func(t *testing.T) {
		db := DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			SSLMode:  "disable",
		}

		assert.Equal(t, "localhost", db.Host)
		assert.Equal(t, 5432, db.Port)
		assert.Equal(t, "testuser", db.User)
		assert.Equal(t, "testpass", db.Password)
		assert.Equal(t, "testdb", db.Name)
		assert.Equal(t, "disable", db.SSLMode)
	})

	t.Run("database config with empty fields", func(t *testing.T) {
		db := DatabaseConfig{}

		assert.Empty(t, db.Host)
		assert.Zero(t, db.Port)
		assert.Empty(t, db.User)
		assert.Empty(t, db.Password)
		assert.Empty(t, db.Name)
		assert.Empty(t, db.SSLMode)
	})
}

func TestServerConfig_Structure(t *testing.T) {
	t.Run("server config fields", func(t *testing.T) {
		server := ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		}

		assert.Equal(t, "0.0.0.0", server.Host)
		assert.Equal(t, 8080, server.Port)
	})

	t.Run("server config with empty fields", func(t *testing.T) {
		server := ServerConfig{}

		assert.Empty(t, server.Host)
		assert.Zero(t, server.Port)
	})
}

func TestConfig_Structure(t *testing.T) {
	t.Run("complete configuration", func(t *testing.T) {
		cfg := &Config{
			Database: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Name:     "db",
				SSLMode:  "disable",
			},
			Server: ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
			LogLevel: "info",
		}

		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "user", cfg.Database.User)
		assert.Equal(t, "pass", cfg.Database.Password)
		assert.Equal(t, "db", cfg.Database.Name)
		assert.Equal(t, "disable", cfg.Database.SSLMode)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("empty configuration", func(t *testing.T) {
		cfg := &Config{}

		assert.Empty(t, cfg.Database.Host)
		assert.Zero(t, cfg.Database.Port)
		assert.Empty(t, cfg.Database.User)
		assert.Empty(t, cfg.Database.Password)
		assert.Empty(t, cfg.Database.Name)
		assert.Empty(t, cfg.Database.SSLMode)
		assert.Empty(t, cfg.Server.Host)
		assert.Zero(t, cfg.Server.Port)
		assert.Empty(t, cfg.LogLevel)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("get existing environment variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test-value")
		result := getEnv("TEST_VAR", "default")
		assert.Equal(t, "test-value", result)
		os.Unsetenv("TEST_VAR")
	})

	t.Run("get non-existent environment variable returns default", func(t *testing.T) {
		result := getEnv("NON_EXISTENT", "default-value")
		assert.Equal(t, "default-value", result)
	})

	t.Run("get empty environment variable returns default", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		result := getEnv("EMPTY_VAR", "default")
		assert.Equal(t, "default", result)
		os.Unsetenv("EMPTY_VAR")
	})

	t.Run("get environment variable with spaces", func(t *testing.T) {
		os.Setenv("SPACE_VAR", "value with spaces")
		result := getEnv("SPACE_VAR", "default")
		assert.Equal(t, "value with spaces", result)
		os.Unsetenv("SPACE_VAR")
	})
}

func TestLoadWithInvalidPorts(t *testing.T) {
	// This test would cause the program to exit due to log.Fatal
	// In a real scenario, you might want to refactor the Load function
	// to return errors instead of calling log.Fatal
	t.Run("test port parsing behavior", func(t *testing.T) {
		// We can't easily test log.Fatal behavior in unit tests
		// This is a limitation of the current implementation
		// In production code, you'd want to return errors instead

		// Test that valid ports work
		os.Setenv("DB_PORT", "5432")
		os.Setenv("SERVER_PORT", "8080")

		cfg := Load()
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, 8080, cfg.Server.Port)

		clearEnvVars()
	})
}

// Test environment loading from different sources
func TestEnvironmentLoading(t *testing.T) {
	t.Run("test with all valid environment variables", func(t *testing.T) {
		clearEnvVars()

		// Set all possible environment variables
		envVars := map[string]string{
			"DB_HOST":     "test-host",
			"DB_PORT":     "3306",
			"DB_USER":     "test-user",
			"DB_PASSWORD": "test-password",
			"DB_NAME":     "test-database",
			"DB_SSLMODE":  "require",
			"SERVER_HOST": "test-server",
			"SERVER_PORT": "9000",
			"LOG_LEVEL":   "debug",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
		}

		cfg := Load()

		assert.Equal(t, "test-host", cfg.Database.Host)
		assert.Equal(t, 3306, cfg.Database.Port)
		assert.Equal(t, "test-user", cfg.Database.User)
		assert.Equal(t, "test-password", cfg.Database.Password)
		assert.Equal(t, "test-database", cfg.Database.Name)
		assert.Equal(t, "require", cfg.Database.SSLMode)
		assert.Equal(t, "test-server", cfg.Server.Host)
		assert.Equal(t, 9000, cfg.Server.Port)
		assert.Equal(t, "debug", cfg.LogLevel)

		clearEnvVars()
	})
}

// Helper function to clear environment variables
func clearEnvVars() {
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"SERVER_HOST", "SERVER_PORT", "LOG_LEVEL",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
