package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/pterm/pterm"
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

// update items table
func updateTable(items []zabbixmonItem) pterm.TableData {
	table := pterm.TableData{{"Host", "Status", "Description", "Ack", "URL"}}
	for _, item := range items {
		table = append(table, []string{item.Host, item.Status, item.Description, fmt.Sprintf("%t", item.Ack), item.Url})
	}

	return table
}

// send notification for all items
func notify(items []zabbixmonItem) {
	for _, item := range items {
		log.Debug().Str("type", "new_item").Str("item", fmt.Sprintf("%#v", item)).Send()
		if err := beeep.Notify(fmt.Sprintf("%s - %s", item.Status, item.Host), item.Description, "assets/information.png"); err != nil {
			log.Error().Err(err).Str("scope", "sending notification").Send()
			os.Exit(1)
		}
	}
}

func dumpJsonIfRedirect(items []zabbixmonItem) {
	o, _ := os.Stdout.Stat()
	if (o.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
		if data, err := json.Marshal(items); err != nil {
			log.Error().Err(err).Str("scope", "dumping json").Send()
			os.Exit(1)
		} else {
			fmt.Println(string(data))
			os.Exit(0)
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	cfg := config

	// set log level
	logLevel := lo.Ternary(cfg.Debug, zerolog.DebugLevel, zerolog.InfoLevel)
	zerolog.SetGlobalLevel(logLevel)

	// dump settings in logs
	log.Debug().Str("type", "settings").Str("settings", fmt.Sprintf("%#v", cfg)).Send()

	var items []zabbixmonItem
	var prevItems []zabbixmonItem

	// zabbix auth
	zapi := getSession(cfg.Server, cfg.Username, cfg.Password)

	// dump json if output is redirected
	dumpJsonIfRedirect(getItems(zapi, cfg.ItemTypes, cfg.MinSeverity, cfg.Grep))

	// setup ui
	area, _ := pterm.DefaultArea.Start()

	for {
		// backup items to detect changes
		if items != nil {
			prevItems = append([]zabbixmonItem(nil), items...)
		}

		// fetch items
		items = getItems(zapi, cfg.ItemTypes, cfg.MinSeverity, cfg.Grep)

		// update table data
		table := updateTable(items)

		// update ui
		data, err := pterm.DefaultTable.WithHasHeader().WithData(table).Srender()
		if err != nil {
			log.Error().Err(err).Str("scope", "rendering table").Send()
			os.Exit(1)
		}
		area.Update(data)

		// detect changes and send notification
		if cfg.Notify && prevItems != nil {
			newItems, _ := lo.Difference(items, prevItems)
			notify(newItems)
		}

		// wait
		time.Sleep(time.Duration(cfg.Refresh) * time.Second)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Str("scope", "command execution").Send()
		os.Exit(1)
	}
}
