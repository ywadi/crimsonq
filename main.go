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
	crimsonQ := Structs.S_GOQ{}
	viperq.Init()
	Logger.Init()
	Servers.InitCommands()
	go Servers.HTTP_Start(&crimsonQ)
	Servers.StartRedCon(":"+viper.GetString("crimson_settings.port"), &crimsonQ)
}
