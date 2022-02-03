package main

import (
	"ywadi/goq/RedconQ"
	"ywadi/goq/Structs"
	viperq "ywadi/goq/viperQ"

	"github.com/spf13/viper"
)

func main() {
	viperq.Init()
	crimsonQ := Structs.S_GOQ{}
	RedconQ.StartRedCon(":"+viper.GetString("redcon_settings.port"), &crimsonQ)
}
