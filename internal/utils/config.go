package utils

import (
	"flag"
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress       string `mapstructure:"RUN_ADDRESS"`         // Адрес и порт запуска сервиса
	DBSource            string `mapstructure:"DATABASE_URI"`              // Адрес подключения к базе данных
	AccrualSystemAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"` // Адрес системы расчёта начислений
}

// type AccrualOrderResponse struct {
//     Order   string  `json:"order"`
//     Status  string  `json:"status"`
//     Accrual float64 `json:"accrual,omitempty"`
// }

func LoadConfig(path string) (config Config, err error) {
	// Чтение переменных окружения
	viper.AutomaticEnv()

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