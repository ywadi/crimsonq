package main

import (
	"ywadi/crimsonq/Logger"
	"ywadi/crimsonq/RedconQ"
	"ywadi/crimsonq/Structs"
	viperq "ywadi/crimsonq/viperQ"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func main() {
	viperq.Init()
	Logger.Init()
	log.Info("Starting...")
	crimsonQ := Structs.S_GOQ{}
	RedconQ.StartRedCon(":"+viper.GetString("crimson_settings.port"), &crimsonQ)
}
