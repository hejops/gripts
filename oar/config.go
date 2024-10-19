package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	MaxDays int `mapstructure:"max_days"`
	Dest    string

	Bandcamp struct {
		Username string
	}
	Youtube struct {
		Urls []string
	}
}

// var Cfg = LoadConfig()
var Cfg *Config

func InitConfig() {
	// required
	viper.Set("bandcamp.username", readline("Bandcamp username"))

	viper.SetDefault("max_days", 7)
	viper.SetDefault("youtube.urls", []string{})

	err := viper.SafeWriteConfigAs("config.toml")
	if err != nil {
		panic(err)
	}

	fmt.Println("Wrote new config: config.toml")
}

func LoadConfig() {
	// os.Remove("config.toml")

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No config found, creating...")
		InitConfig()
	}

	// var c Config
	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	// fmt.Println(c)
	// return c
}
