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
	Telemetry   Telemetry         `mapstructure:"telemetry"`
	Auth        AuthConfig        `mapstructure:"auth"`
}

type ApplicationConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
}

type Telemetry struct {
	SentryDSN string `mapstructure:"sentry_dsn"`
}

type AuthConfig struct {
	Auth0Domain       string `mapstructure:"auth_0_domain"`
	Auth0ClientId     string `mapstructure:"auth_0_client_id"`
	Auth0ClientSecret string `mapstructure:"auth_0_client_secret"`
	Auth0CallbackURL  string `mapstructure:"auth_0_callback_url"`
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
	if err := viper.BindEnv("database.connection_string", "DATABASE_URL"); err != nil {
		return nil, fmt.Errorf("Failed to bind DATABASE_URL")
	}
	if err := viper.BindEnv("telemetry.sentry_dsn", "TELEMETRY_SENTRY_DSN"); err != nil {
		return nil, fmt.Errorf("Failed to bind TELEMETRY_SENTRY_DSN")
	}
	if err := viper.BindEnv("auth.auth_0_domain", "AUTH0_DOMAIN"); err != nil {
		return nil, fmt.Errorf("Failed to bind AUTH0_DOMAIN")
	}
	if err := viper.BindEnv("auth.auth_0_client_id", "AUTH0_CLIENT_ID"); err != nil {
		return nil, fmt.Errorf("Failed to bind AUTH0_CLIENT_ID")
	}
	if err := viper.BindEnv("auth.auth_0_client_secret", "AUTH0_CLIENT_SECRET"); err != nil {
		return nil, fmt.Errorf("Failed to bind AUTH0_CLIENT_SECRET")
	}
	if err := viper.BindEnv("auth.auth_0_callback_url", "AUTH0_CALLBACK_URL"); err != nil {
		return nil, fmt.Errorf("Failed to bind AUTH0_CALLBACK_URL")
	}

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
