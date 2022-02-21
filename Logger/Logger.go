package Logger

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init() {
	f, err := os.OpenFile(viper.GetString("crimson_settings.log_file"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(f)
}
