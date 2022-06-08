package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/markkurossi/tabulate"
	"github.com/nikaro/zabbixmon/api"
	"github.com/nikaro/zabbixmon/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zabbixmon",
	Short: "Zabbix Status Monitoring",
	Long:  ``,
	Run:   run,
}

// initialize command
func init() {
	cobra.OnInitialize(config.InitConfig)

	// set flags
	rootCmd.Flags().StringP("server", "s", "", "zabbix server url")
	rootCmd.Flags().StringP("username", "u", "", "zabbix username")
	rootCmd.Flags().StringP("password", "p", "", "zabbix password")
	rootCmd.Flags().IntP("refresh", "r", 0, "data refreshing interval")
	rootCmd.Flags().BoolP("notify", "n", false, "enable notifications")
	rootCmd.Flags().StringP("min-severity", "m", "", "minimum trigger severity")
	rootCmd.Flags().StringSliceP("item-types", "i", nil, "items state types")
	rootCmd.Flags().StringP("log-level", "l", "", "logging level")

	// bind flag to config
	viper.BindPFlag("server", rootCmd.Flags().Lookup("server"))
	viper.BindPFlag("username", rootCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	viper.BindPFlag("refresh", rootCmd.Flags().Lookup("refresh"))
	viper.BindPFlag("notify", rootCmd.Flags().Lookup("notify"))
	viper.BindPFlag("min-severity", rootCmd.Flags().Lookup("min-severity"))
	viper.BindPFlag("item-types", rootCmd.Flags().Lookup("item-types"))
	viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level"))
}

// check and set global log level
func setLogLevel(logLevel string) {
	var logLevels = map[string]zerolog.Level{
		"panic": zerolog.PanicLevel,
		"fatal": zerolog.FatalLevel,
		"error": zerolog.ErrorLevel,
		"warn":  zerolog.WarnLevel,
		"info":  zerolog.InfoLevel,
		"debug": zerolog.DebugLevel,
		"trace": zerolog.TraceLevel,
	}

	// check log level
	logLevelsKeys := lo.Keys[string, zerolog.Level](logLevels)
	if _, found := lo.Find[string](logLevelsKeys, func(i string) bool {
		return i == logLevel
	}); !found {
		err := errors.New(fmt.Sprintf("unknown log level '%s', not in %v", logLevel, logLevelsKeys))
		log.Error().Err(err).Send()
		os.Exit(1)
	}

	// set log level
	zerolog.SetGlobalLevel(logLevels[logLevel])
}

// build items table
func buildTable(items []api.Item) (table *tabulate.Tabulate) {
	table = tabulate.New(tabulate.Unicode)
	table.Header("Host")
	table.Header("Status")
	table.Header("Description")
	table.Header("Ack")
	table.Header("URL")
	lo.ForEach[api.Item](items, func(x api.Item, _ int) {
		row := table.Row()
		row.Column(x.Host)
		row.Column(x.Status)
		row.Column(x.Description)
		row.Column(fmt.Sprintf("%t", x.Ack))
		row.Column(x.Url)
	})

	return table
}

// send notification for all items
func notify(items []api.Item) {
	for _, item := range items {
		log.Debug().Str("type", "new_item").Str("item", fmt.Sprintf("%#v", item)).Send()
		err := beeep.Notify(fmt.Sprintf("%s - %s", item.Status, item.Host), item.Description, "assets/information.png")
		if err != nil {
			log.Error().Err(err).Send()
			os.Exit(1)
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	var items []api.Item
	var prevItems []api.Item
	cfg := config.Config

	// set log level
	logLevel := cfg.LogLevel
	setLogLevel(logLevel)

	// dump settings in logs
	log.Debug().Str("type", "settings").Str("settings", fmt.Sprintf("%#v", cfg)).Send()

	// zabbix auth
	zapi := api.GetSession(cfg.Server, cfg.Username, cfg.Password)

	// catch ctrl+c signal
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	for {
		// backup items to detect changes
		if items != nil {
			prevItems = append([]api.Item(nil), items...)
		}

		// fetch items
		items = api.GetItems(zapi, cfg.ItemTypes, cfg.MinSeverity)

		// build table
		table := buildTable(items)

		// dump json if output is redirected
		o, _ := os.Stdout.Stat()
		if (o.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
			if data, err := json.Marshal(table); err != nil {
				log.Error().Err(err).Send()
				os.Exit(1)
			} else {
				fmt.Println(string(data))
				return
			}
		}

		// clear terminal
		cmd := lo.Ternary[*exec.Cmd](runtime.GOOS == "windows", exec.Command("cmd", "/c", "cls"), exec.Command("clear"))
		cmd.Stdout = os.Stdout
		cmd.Run()

		// print table
		table.Print(os.Stdout)

		// detect changes and send notification
		if cfg.Notify && prevItems != nil {
			newItems, _ := lo.Difference[api.Item](items, prevItems)
			notify(newItems)
		}

		// wait
		time.Sleep(time.Duration(cfg.Refresh) * time.Second)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
