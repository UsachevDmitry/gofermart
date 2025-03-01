package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBSource string `mapstructure:"DB_SOURCE"`
	serverAddress string `mapstructure:"SERVER_ADDRESS"`
}
//LoadConfig func reads confihureations from file or environment variables
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// read environmentak variables
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err!= nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}