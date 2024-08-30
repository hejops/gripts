package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Bandcamp struct {
		Username string
	}
}

func InitConfig() {
	viper.Set(
		"bandcamp.username",
		readline("Bandcamp username"),
	)

	err := viper.SafeWriteConfigAs("config.toml")
	if err != nil {
		panic(err)
	}

	fmt.Println("Wrote new config: config.toml")
}

func LoadConfig() Config {
	// os.Remove("config.toml")

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No config found, creating...")
		InitConfig()
	}

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	// fmt.Println(c)
	return c
}
