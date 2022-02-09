package viperq

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("crimson.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatal(fmt.Errorf("fatal error config file: %w \n", err))
	}
	log.Info("Settings loaded...")
	log.WithField("settings", viper.GetStringMap("crimson_settings")).Info("Loaded settings")
}
