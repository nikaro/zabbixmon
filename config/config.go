package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ZabbixMonConfig struct {
	ConfigFile  string
	Server      string
	Username    string
	Password    string
	LogLevel    string
	ItemTypes   []string
	MinSeverity string
	Refresh     int
	Notify      bool
}

var Config *ZabbixMonConfig

func InitConfig() {
	viper.SetConfigName("config")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}

	// search paths
	viper.AddConfigPath("/etc/zabbixmon")
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		viper.AddConfigPath("$XDG_CONFIG_HOME/zabbixmon")
	}
	viper.AddConfigPath(home + "/.config/zabbixmon")
	viper.AddConfigPath(home + "/.zabbixmon")
	viper.AddConfigPath(".")

	// set defaults
	viper.SetDefault("log-level", "info")
	viper.SetDefault("item-types", []string{"down", "unack", "ack", "unknown"})
	viper.SetDefault("min-severity", "average")
	viper.SetDefault("refresh", 60)
	viper.SetDefault("notify", false)

	// bind environment variables
	viper.SetEnvPrefix("zxmon")
	viper.AutomaticEnv()

	// read config
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}

	// check mandatory values
	for _, setting := range []string{"server", "username", "password"} {
		if !viper.IsSet(setting) {
			log.Error().Str("scope", "config").Msg(fmt.Sprintf("'%s' is not set", setting))
			os.Exit(1)
		}
	}

	// update global config object
	Config = &ZabbixMonConfig{
		ConfigFile:  viper.ConfigFileUsed(),
		Server:      viper.GetString("server"),
		Username:    viper.GetString("username"),
		Password:    viper.GetString("password"),
		LogLevel:    viper.GetString("log-level"),
		ItemTypes:   viper.GetStringSlice("item-types"),
		MinSeverity: viper.GetString("min-severity"),
		Refresh:     viper.GetInt("refresh"),
		Notify:      viper.GetBool("notify"),
	}
}
