package utils

import (
	"flag"
	"github.com/spf13/viper"
)

// type Config struct {
// 	DBSource string `mapstructure:"DB_SOURCE"`
// 	serverAddress string `mapstructure:"SERVER_ADDRESS"`
// }
// //LoadConfig func reads confihureations from file or environment variables
// func LoadConfig(path string) (config Config, err error) {
// 	viper.AddConfigPath(path)
// 	viper.SetConfigName("app")
// 	viper.SetConfigType("env")

// 	// read environmentak variables
// 	viper.AutomaticEnv()

// 	err = viper.ReadInConfig()
// 	if err!= nil {
// 		return
// 	}
// 	err = viper.Unmarshal(&config)
// 	return
// }

type Config struct {
	DBSource            string `mapstructure:"DB_SOURCE"`              // Адрес подключения к базе данных
	ServerAddress       string `mapstructure:"SERVER_ADDRESS"`         // Адрес и порт запуска сервиса
	AccrualSystemAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"` // Адрес системы расчёта начислений
}

func LoadConfig(path string) (config Config, err error) {
	// Чтение конфигурации из файла
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// Чтение переменных окружения
	viper.AutomaticEnv()

	// Парсинг флагов
	flag.StringVar(&config.ServerAddress, "a", "", "Адрес и порт запуска сервиса")
	flag.StringVar(&config.DBSource, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	flag.Parse()

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

	// Переопределение значений флагами, если они заданы
	if config.ServerAddress == "" {
		config.ServerAddress = viper.GetString("SERVER_ADDRESS")
	}
	if config.DBSource == "" {
		config.DBSource = viper.GetString("DB_SOURCE")
	}
	if config.AccrualSystemAddress == "" {
		config.AccrualSystemAddress = viper.GetString("ACCRUAL_SYSTEM_ADDRESS")
	}

	return
}