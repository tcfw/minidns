package main

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("forwarders", []string{"1.1.1.1", "1.0.0.1"})
	viper.SetDefault("port", 53)
}
