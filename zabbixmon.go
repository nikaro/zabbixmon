package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "zabbixmon",
	Short: "Zabbix Status Monitoring",
	Run:   run,
}

type zabbixmonConfig struct {
	ConfigFile  string
	Server      string
	Username    string
	Password    string
	Debug       bool
	ItemTypes   []string
	MinSeverity string
	Refresh     int
	Notify      bool
	Grep        string
}

var config *zabbixmonConfig

func init() {
	cobra.OnInitialize(initConfig)

	// set flags
	rootCmd.Flags().StringP("server", "s", "", "zabbix server url")
	rootCmd.Flags().StringP("username", "u", "", "zabbix username")
	rootCmd.Flags().StringP("password", "p", "", "zabbix password")
	rootCmd.Flags().IntP("refresh", "r", 0, "data refreshing interval")
	rootCmd.Flags().BoolP("notify", "n", false, "enable notifications")
	rootCmd.Flags().StringP("min-severity", "m", "", "minimum trigger severity")
	rootCmd.Flags().StringSliceP("item-types", "i", nil, "items state types")
	rootCmd.Flags().BoolP("debug", "d", false, "enable debug logs")
	rootCmd.Flags().StringP("grep", "g", "", "regexp to filter items on hostname")

	// bind flag to config
	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		log.Warn().Err(err).Send()
	}
}

func initConfig() {
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
	viper.SetDefault("debug", false)
	viper.SetDefault("item-types", []string{"down", "unack", "ack", "unknown"})
	viper.SetDefault("min-severity", "average")
	viper.SetDefault("refresh", 60)
	viper.SetDefault("notify", false)
	viper.SetDefault("grep", "")

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
	config = &zabbixmonConfig{
		ConfigFile:  viper.ConfigFileUsed(),
		Server:      viper.GetString("server"),
		Username:    viper.GetString("username"),
		Password:    viper.GetString("password"),
		Debug:       viper.GetBool("debug"),
		ItemTypes:   viper.GetStringSlice("item-types"),
		MinSeverity: viper.GetString("min-severity"),
		Refresh:     viper.GetInt("refresh"),
		Notify:      viper.GetBool("notify"),
		Grep:        viper.GetString("grep"),
	}
}

func run(cmd *cobra.Command, args []string) {
	// set log level
	logLevel := lo.Ternary(config.Debug, zerolog.DebugLevel, zerolog.InfoLevel)
	zerolog.SetGlobalLevel(logLevel)

	// dump settings in logs
	log.Debug().Str("type", "settings").Str("settings", fmt.Sprintf("%#v", config)).Send()

	// intialize model
	m := initModel()

	// dump json if output is redirected
	dumpJsonIfRedirect(getItems(m.zapi, config.ItemTypes, config.MinSeverity, config.Grep))

	// start ui
	if err := tea.NewProgram(m).Start(); err != nil {
		log.Error().Err(err).Str("scope", "starting program").Send()
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Str("scope", "command execution").Send()
		os.Exit(1)
	}
}
