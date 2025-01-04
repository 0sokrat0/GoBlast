package configs

import (
	"GoBlast/pkg/logger"
	"fmt"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Broker       NATSConfig         `mapstructure:"broker"`
	Encricrypted EncricryptedConfig `mapstructure:"encrypted"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	JWTSecret   string `mapstructure:"jwt_secret"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SslMode  string `mapstructure:"sslmode"`
}

type NATSConfig struct {
	URL string `mapstructure:"url"`
}

type EncricryptedConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

var AppConfigInstance *Config

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config") // Ожидается файл config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	viper.SetEnvPrefix("goblast")

	if err := viper.ReadInConfig(); err != nil {
		logger.Log.Error("Ошибка чтения файла конфигурации", zap.Error(err))
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		logger.Log.Error("Ошибка десериализации конфигурации", zap.Error(err))
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	logger.Log.Info("Конфигурация успешно загружена", zap.String("path", path))
	AppConfigInstance = &config
	return &config, nil
}

func GetDSN(config DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Name, config.SslMode)
}
