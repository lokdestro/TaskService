package config

import (
	"TaskService/pkg/kafka"
	"TaskService/pkg/logger"
	"fmt"

	"TaskService/internal/storage/postgres"

	"github.com/spf13/viper"
)

func Psql() postgres.Config {
	result := postgres.Config{
		URL: fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
			viper.GetString("postgres.user"),
			viper.GetString("postgres.password"),
			viper.GetString("postgres.url"),
			viper.GetString("postgres.name"),
		),
		Driver: viper.GetString("postgres.driver"),
	}

	return result
}

func Log() logger.Config {
	return logger.Config{
		Dir:               viper.GetString("logger.dir"),
		Filename:          viper.GetString("logger.filename"),
		Level:             viper.GetString("logger.level"),
		MaxSizeMB:         viper.GetInt("logger.max_size_mb"),
		MaxBackups:        viper.GetInt("logger.max_backups"),
		MaxAgeDays:        viper.GetInt("logger.max_age_days"),
		Compress:          viper.GetBool("logger.compress"),
		DuplicateToStdout: viper.GetBool("logger.duplicate_to_stdout"),
		TimeFormat:        viper.GetString("logger.time_format"),
		ServiceName:       viper.GetString("logger.service_name"),
	}
}

func Kfk() kafka.Config {
	result := kafka.Config{
		Broker: viper.GetString("kafka.brokers"),
		Topic:  viper.GetString("kafka.topic"),
	}

	return result
}

func Srv() string {
	return fmt.Sprintf(":%d", viper.GetInt("server.port"))
}
