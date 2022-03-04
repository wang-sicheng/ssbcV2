package config

import (
	viper2 "github.com/spf13/viper"
	"github.com/ssbcV2/global"
)

func Get(key string) interface{}  {
	viper := viper2.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(global.RootDir + "/config/")

	if err := viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}
	return viper.Get(key)
}
