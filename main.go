package main

import (
	"strconv"
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
	enableHTTP, err := strconv.ParseBool(viper.GetString("HTTP.enabled"))

	if err != nil {
		log.Fatal("crimson.config value for HTTP enabled needs to be true or false")
	}
	if enableHTTP {
		go Servers.HTTP_Start(&crimsonQ)
	}
	Servers.StartRedCon(":"+viper.GetString("RESP.port"), &crimsonQ)
}
