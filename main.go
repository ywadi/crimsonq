package main

import (
	"ywadi/crimsonq/Logger"
	"ywadi/crimsonq/Servers"
	"ywadi/crimsonq/Structs"
	viperq "ywadi/crimsonq/viperQ"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func main() {
	log.Info("Starting...")
	viperq.Init()
	Logger.Init()
	Servers.InitCommands()
	crimsonQ := Structs.S_GOQ{}
	Servers.StartRedCon(":"+viper.GetString("crimson_settings.port"), &crimsonQ)
}
