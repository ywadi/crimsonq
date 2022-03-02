package viperq

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("crimson.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatal(fmt.Errorf("fatal error config file: %w \n", err))
	}
	log.Info("Settings loaded...")
	log.WithField("settings", viper.GetStringMap("crimson_settings")).Info("Loaded settings")
	httpUsernameVal, httpUserPresent := os.LookupEnv("CRIMSONQ_HTTP_USER")
	httpPassVal, httpPassPresent := os.LookupEnv("CRIMSONQ_HTTP_PASS")
	respPassVal, respPassPresent := os.LookupEnv("CRIMSONQ_RESP_PASS")
	if !httpUserPresent || !httpPassPresent {
		err := os.Setenv("CRIMSONQ_HTTP_USER", viper.GetString("HTTP.username"))
		if err != nil {
			log.Fatal(err)
		}
		err = os.Setenv("CRIMSONQ_HTTP_PASS", viper.GetString("HTTP.password"))
		if err != nil {
			log.Fatal(err)
		}
	}
	if !respPassPresent {
		err := os.Setenv("CRIMSONQ_RESP_PASS", viper.GetString("RESP.password"))
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Info("HTTP USERNAME", httpUserPresent, httpUsernameVal)
	log.Info("HTTP PASS", httpPassPresent, httpPassVal)
	log.Info("RESP PASS", respPassPresent, respPassVal)
}
