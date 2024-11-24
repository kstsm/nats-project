package configs

import (
	"github.com/spf13/viper"
	"log"
)

var Config struct {
	SubscriberAddr string
}

func init() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Ошибка при чтении конфигурационного файла: %s", err)
	}

	Config.SubscriberAddr = viper.GetString("SUBSCRIBER_ADDR")
}
