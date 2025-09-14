package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Application ApplicationConfig `mapstructure:"application"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Telemetry Telemetry `mapstructure:"telemetry"`
}

type ApplicationConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
}

type Telemetry struct {
	SentryDSN string `mapstructure:"sentry_dsn"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get the environment
	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "dev"
	}

	// Setup Viper
	viper.SetConfigName(env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Read the base configuration
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	// Configure environment variable handling
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "_"))
	viper.AutomaticEnv()

	// Bind specific environment variables that don't follow the APP_ prefix pattern
	// or have different naming conventions
	viper.BindEnv("database.connection_string", "DATABASE_URL")
	viper.BindEnv("telemetry.sentry_dsn", "TELEMETRY_SENTRY_DSN")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if envVar := os.Getenv("APP_ENVIRONMENT"); envVar != "" {
		config.Application.Environment = envVar
	}

	return &config, nil
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Application.Host, c.Application.Port)
}
