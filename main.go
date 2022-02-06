package main

import (
	"ywadi/crimsonq/RedconQ"
	"ywadi/crimsonq/Structs"
	viperq "ywadi/crimsonq/viperQ"

	"github.com/spf13/viper"
)

func main() {
	viperq.Init()
	crimsonQ := Structs.S_GOQ{}
	RedconQ.StartRedCon(":"+viper.GetString("crimson_settings.port"), &crimsonQ)
}
