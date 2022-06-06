package cmd

import (
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
	var configFile string

	// set config file flags
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")

	// set config file
	if configFile != "" {
		viper.SetConfigName(configFile)
	} else {
		viper.SetConfigName("config")

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// search paths
		viper.AddConfigPath("/etc/zabbixmon")
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			viper.AddConfigPath("$XDG_CONFIG_HOME/zabbixmon")
		}
		viper.AddConfigPath(home + "/.config/zabbixmon")
		viper.AddConfigPath(home + "/.zabbixmon")
		viper.AddConfigPath(".")
	}

	// set defaults
	viper.SetDefault("log-level", "info")
	viper.SetDefault("item-types", []string{"down", "unack", "ack", "unknown"})
	viper.SetDefault("min-severity", "average")
	viper.SetDefault("refresh", 60)
	viper.SetDefault("notify", false)

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

	// bind environment variables
	viper.SetEnvPrefix("zxmon")
	viper.AutomaticEnv()

	// read config
	viper.ReadInConfig()
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
	if _, found := lo.Find[string](lo.Keys[string, zerolog.Level](logLevels), func(i string) bool {
		return i == logLevel
	}); !found {
		panic(fmt.Sprintf("unknown log level '%s'", logLevel))
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
		row.Column(fmt.Sprintf("%v", x.Ack))
		row.Column(x.Url)
	})

	return table
}

// send notification for all items
func notify(items []api.Item) {
	for _, item := range items {
		log.Debug().Str("type", "new_item").Str("item", fmt.Sprintf("%v", item)).Send()
		err := beeep.Notify(fmt.Sprintf("%s - %s", item.Status, item.Host), item.Description, "assets/information.png")
		if err != nil {
			panic(err)
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	var items []api.Item
	var prevItems []api.Item

	// set log level
	logLevel := viper.GetString("log-level")
	setLogLevel(logLevel)

	// dump settings in logs
	log.Debug().Str("type", "settings").Str("config_file", viper.ConfigFileUsed()).Send()
	log.Debug().Str("type", "settings").Str("settings", fmt.Sprintf("%v", viper.AllSettings())).Send()

	// zabbix auth
	zapi := api.GetSession()

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
		items = api.GetItems(zapi)

		// build table
		table := buildTable(items)

		// clear terminal
		cmd := lo.Ternary[*exec.Cmd](runtime.GOOS == "windows", exec.Command("cmd", "/c", "cls"), exec.Command("clear"))
		cmd.Stdout = os.Stdout
		cmd.Run()

		// print table
		table.Print(os.Stdout)

		// detect changes and send notification
		if viper.GetBool("notify") && prevItems != nil {
			newItems, _ := lo.Difference[api.Item](items, prevItems)
			notify(newItems)
		}

		// wait
		time.Sleep(time.Duration(viper.GetInt("refresh")) * time.Second)
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
