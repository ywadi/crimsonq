package viperq

import (
	"fmt"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("crimson.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.crimsonQ")
	viper.AddConfigPath("./defaultSettings")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	fmt.Println("Settings loaded...")
	fmt.Println(viper.AllKeys(), ", loaded from settings file ", viper.ConfigFileUsed())
}
