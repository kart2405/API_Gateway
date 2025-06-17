package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var RouteMap map[string]string

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	routes := viper.GetStringMapString("routes")
	RouteMap = routes
	fmt.Printf("Loaded routes: %+v\n", RouteMap)
	return nil
}
