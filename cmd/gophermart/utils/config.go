package utils

import (
	"flag"
	"github.com/spf13/viper"
)

type Config struct {
	DBSource            string `mapstructure:"DATABASE_URI"`              // Адрес подключения к базе данных
	ServerAddress       string `mapstructure:"RUN_ADDRESS"`         // Адрес и порт запуска сервиса
	AccrualSystemAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"` // Адрес системы расчёта начислений
}

func LoadConfig(path string) (config Config, err error) {
	// Чтение конфигурации из файла
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// Чтение переменных окружения
	viper.AutomaticEnv()

	// Чтение конфигурации из файла
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// Привязка значений из файла и переменных окружения к структуре Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	// Парсинг флагов
	flag.StringVar(&config.ServerAddress, "a", "", "Адрес и порт запуска сервиса")
	flag.StringVar(&config.DBSource, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	flag.Parse()

	// Переопределение значений флагами, если они заданы
	if config.ServerAddress == "" {
		config.ServerAddress = viper.GetString("RUN_ADDRESS")
	}
	if config.DBSource == "" {
		config.DBSource = viper.GetString("DATABASE_URI")
	}
	if config.AccrualSystemAddress == "" {
		config.AccrualSystemAddress = viper.GetString("ACCRUAL_SYSTEM_ADDRESS")
	}

	return
}