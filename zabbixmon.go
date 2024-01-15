package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
	ConfigFile     string
	Server         string
	ServerInsecure bool
	Username       string
	Password       string
	Debug          bool
	ItemTypes      []string
	MinSeverity    string
	Refresh        int
	Notify         bool
	Grep           string
}

var config *zabbixmonConfig

func init() {
	cobra.OnInitialize(initConfig)

	// set flags
	rootCmd.Flags().StringP("server", "s", "", "zabbix server url")
	rootCmd.Flags().BoolP("server-insecure", "k", false, "do not check tls certificate")
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
		slog.Warn(err.Error())
	}
}

func initConfig() {
	viper.SetConfigName("config")

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error(err.Error())
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
	viper.SetDefault("server-insecure", false)
	viper.SetDefault("grep", "")

	// bind environment variables
	viper.SetEnvPrefix("zxmon")
	viper.AutomaticEnv()

	// read config
	if err := viper.ReadInConfig(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// check mandatory values
	for _, setting := range []string{"server", "username", "password"} {
		if !viper.IsSet(setting) {
			slog.Error(fmt.Sprintf("'%s' is not set", setting), slog.String("scope", "config"))
			os.Exit(1)
		}
	}

	// update global config object
	config = &zabbixmonConfig{
		ConfigFile:     viper.ConfigFileUsed(),
		Server:         viper.GetString("server"),
		ServerInsecure: viper.GetBool("server-insecure"),
		Username:       viper.GetString("username"),
		Password:       viper.GetString("password"),
		Debug:          viper.GetBool("debug"),
		ItemTypes:      viper.GetStringSlice("item-types"),
		MinSeverity:    viper.GetString("min-severity"),
		Refresh:        viper.GetInt("refresh"),
		Notify:         viper.GetBool("notify"),
		Grep:           viper.GetString("grep"),
	}
}

func run(cmd *cobra.Command, args []string) {
	// set log level
	loggerOpts := &slog.HandlerOptions{
		Level: lo.Ternary(config.Debug, slog.LevelDebug, slog.LevelInfo),
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, loggerOpts))
	slog.SetDefault(logger)

	// dump settings in logs
	slog.Debug("", slog.String("type", "settings"), slog.String("settings", fmt.Sprintf("%#v", config)))

	// intialize model
	m := initModel()

	// dump json if output is redirected
	dumpJsonIfRedirect(getItems(m.zapi, config.ItemTypes, config.MinSeverity, config.Grep))

	// start ui
	if _, err := tea.NewProgram(m).Run(); err != nil {
		slog.Error(err.Error(), slog.String("scope", "starting program"))
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error(), slog.String("scope", "command execution"))
		os.Exit(1)
	}
}
