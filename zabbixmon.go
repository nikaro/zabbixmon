package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/cavaliercoder/go-zabbix"
	"github.com/gen2brain/beeep"
	"github.com/markkurossi/tabulate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

var logLevels = map[string]zerolog.Level{
	"panic": zerolog.PanicLevel,
	"fatal": zerolog.FatalLevel,
	"error": zerolog.ErrorLevel,
	"warn":  zerolog.WarnLevel,
	"info":  zerolog.InfoLevel,
	"debug": zerolog.DebugLevel,
	"trace": zerolog.TraceLevel,
}

var triggerSeverity = map[string]int{
	"unknown":     0,
	"information": 1,
	"warning":     2,
	"average":     3,
	"high":        4,
	"critical":    5,
}

var hostAvalability = map[string]int{
	"UNKNOWN":     0,
	"AVAILABLE":   1,
	"UNAVAILABLE": 2,
}

type Item struct {
	Host        string
	Status      string
	Description string
	Ack         bool
	Url         string
}

var items []Item
var prevItems []Item

var command = &cobra.Command{
	Use:   "zabbixmon",
	Short: "Zabbix Status Monitoring",
	Long:  ``,
	Run:   run,
}

func init() {
	// set config file flags
	command.Flags().StringVarP(&configFile, "config", "c", "", "config file")

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
	command.Flags().StringP("server", "s", "", "zabbix server url")
	command.Flags().StringP("username", "u", "", "zabbix username")
	command.Flags().StringP("password", "p", "", "zabbix password")
	command.Flags().IntP("refresh", "r", 0, "data refreshing interval")
	command.Flags().BoolP("notify", "n", false, "enable notifications")
	command.Flags().StringP("min-severity", "m", "", "minimum trigger severity")
	command.Flags().StringSliceP("item-types", "i", nil, "items state types")
	command.Flags().StringP("log-level", "l", "", "logging level")

	// bind flag to config
	viper.BindPFlag("server", command.Flags().Lookup("server"))
	viper.BindPFlag("username", command.Flags().Lookup("username"))
	viper.BindPFlag("password", command.Flags().Lookup("password"))
	viper.BindPFlag("refresh", command.Flags().Lookup("refresh"))
	viper.BindPFlag("notify", command.Flags().Lookup("notify"))
	viper.BindPFlag("min-severity", command.Flags().Lookup("min-severity"))
	viper.BindPFlag("item-types", command.Flags().Lookup("item-types"))
	viper.BindPFlag("log-level", command.Flags().Lookup("log-level"))

	// bind environment variables
	viper.SetEnvPrefix("zxmon")
	viper.AutomaticEnv()

	// read config
	viper.ReadInConfig()
}

func getTriggers(zapi *zabbix.Session) (triggerItemsUnack []Item, triggerItemsAck []Item) {
	// query triggers with unresolved problems
	triggerParams := zabbix.TriggerGetParams{
		GetParameters: zabbix.GetParameters{
			Filter:       map[string]interface{}{"value": 1},
			OutputFields: []string{"triggerid", "description", "priority", "value"},
			SortField:    []string{"priority", "lastchange"},
			SortOrder:    zabbix.SortOrderDescending,
		},
		ActiveOnly:        true,
		MonitoredOnly:     true,
		SelectHosts:       []string{"host"},
		SelectLastEvent:   "extend",
		ExpandDescription: true,
		MinSeverity:       triggerSeverity[viper.GetString("min-severity")],
	}
	triggers, err := zapi.GetTriggers(triggerParams)
	if err != nil {
		panic(err)
	}
	log.Debug().Str("type", "triggers_raw").Str("scope", "all").Str("triggers", fmt.Sprintf("%v", triggers)).Send()

	// ensure we only have currently in problem state
	triggers = lo.Filter[zabbix.Trigger](triggers, func(x zabbix.Trigger, _ int) bool { return x.LastEvent.Value == 1 })

	// tranform triggers into structured items
	severity := lo.Invert[string, int](triggerSeverity)
	triggerItemsAll := lo.Map[zabbix.Trigger, Item](triggers, func(x zabbix.Trigger, _ int) Item {
		return Item{
			Host:        x.Hosts[0].Hostname,
			Status:      strings.ToUpper(severity[x.Severity]),
			Description: x.Description,
			Ack:         x.LastEvent.Acknowledged,
			Url:         fmt.Sprintf("%s/tr_events.php?triggerid=%s&eventid=%s", viper.GetString("server"), x.TriggerID, x.LastEvent.EventID),
		}
	})
	log.Debug().Str("type", "triggers").Str("scope", "all").Str("items", fmt.Sprintf("%v", triggerItemsAll)).Send()

	// filter unacknowledged items
	triggerItemsUnack = lo.Filter[Item](triggerItemsAll, func(x Item, _ int) bool { return !x.Ack })
	log.Debug().Str("type", "triggers").Str("scope", "unack").Str("items", fmt.Sprintf("%v", triggerItemsUnack)).Send()

	// filter acknowledged items
	triggerItemsAck = lo.Filter[Item](triggerItemsAll, func(x Item, _ int) bool { return x.Ack })
	log.Debug().Str("type", "triggers").Str("scope", "ack").Str("items", fmt.Sprintf("%v", triggerItemsAck)).Send()

	return triggerItemsUnack, triggerItemsAck
}

func getHosts(zapi *zabbix.Session) (hostItemsUnavailable []Item, hostItemsUnknown []Item) {
	// query hosts with problems
	hostParams := zabbix.HostGetParams{
		GetParameters: zabbix.GetParameters{
			Filter:       map[string]interface{}{"available": []int{hostAvalability["UNKNOWN"], hostAvalability["UNAVAILABLE"]}},
			OutputFields: []string{"hostid", "host", "available"},
		},
	}
	hosts, err := zapi.GetHosts(hostParams)
	if err != nil {
		panic(err)
	}
	log.Debug().Str("type", "hosts_raw").Str("scope", "all").Str("hosts", fmt.Sprintf("%v", hosts)).Send()

	// tranform hosts into structured items
	availability := lo.Invert[string, int](hostAvalability)
	hostItemsAll := lo.Map[zabbix.Host, Item](hosts, func(x zabbix.Host, _ int) Item {
		return Item{
			Host:        x.Hostname,
			Status:      availability[x.Available],
			Description: fmt.Sprintf("Host in %s state", availability[x.Available]),
			Ack:         false,
			Url:         fmt.Sprintf("%s/hostinventories.php?hostid=%s", viper.GetString("server"), x.HostID),
		}
	})
	log.Debug().Str("type", "hosts").Str("scope", "all").Str("items", fmt.Sprintf("%v", hostItemsAll)).Send()

	// filter unavailable items
	hostItemsUnavailable = lo.Filter[Item](hostItemsAll, func(x Item, _ int) bool { return x.Status == "UNAVAILABLE" })
	log.Debug().Str("type", "hosts").Str("scope", "unavailable").Str("items", fmt.Sprintf("%v", hostItemsUnavailable)).Send()

	// filter unknown items
	hostItemsUnknown = lo.Filter[Item](hostItemsAll, func(x Item, _ int) bool { return x.Status == "UNKNOWN" })
	log.Debug().Str("type", "hosts").Str("scope", "unknown").Str("items", fmt.Sprintf("%v", hostItemsUnknown)).Send()

	return hostItemsUnavailable, hostItemsUnknown
}

func run(cmd *cobra.Command, args []string) {
	// check log level
	logLevel := viper.GetString("log-level")
	if _, found := lo.Find[string](lo.Keys[string, zerolog.Level](logLevels), func(i string) bool {
		return i == logLevel
	}); !found {
		cobra.CheckErr(fmt.Sprintf("unknown log level '%s'", logLevel))
	}
	// set log level
	zerolog.SetGlobalLevel(logLevels[logLevel])

	log.Debug().Str("type", "settings").Str("config_file", viper.ConfigFileUsed()).Send()
	log.Debug().Str("type", "settings").Str("settings", fmt.Sprintf("%v", viper.AllSettings())).Send()

	// authenticate to zabbix server
	zapi, err := zabbix.NewSession(
		viper.GetString("server")+"/api_jsonrpc.php",
		viper.GetString("username"),
		viper.GetString("password"),
	)
	if err != nil {
		cobra.CheckErr(err)
	}

	// catch sigterm (like ctrl+c) signal
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	for {
		// backup items to detect changes
		if items != nil {
			prevItems = append([]Item(nil), items...)
		}
		// reset items
		items = nil

		// get triggers
		triggerItemsUnack, triggerItemsAck := getTriggers(zapi)
		if present := lo.Contains[string](viper.GetStringSlice("item-types"), "unack"); present {
			items = append(items, triggerItemsUnack...)
		}
		if present := lo.Contains[string](viper.GetStringSlice("item-types"), "ack"); present {
			items = append(items, triggerItemsAck...)
		}

		// get hosts
		hostItemsUnavailable, hostItemsUnknown := getHosts(zapi)
		if present := lo.Contains[string](viper.GetStringSlice("item-types"), "down"); present {
			items = append(items, hostItemsUnavailable...)
		}
		if present := lo.Contains[string](viper.GetStringSlice("item-types"), "unknown"); present {
			items = append(items, hostItemsUnknown...)
		}

		// ouput
		table := tabulate.New(tabulate.Unicode)
		table.Header("Host")
		table.Header("Status")
		table.Header("Description")
		table.Header("Ack")
		table.Header("URL")
		lo.ForEach[Item](items, func(x Item, _ int) {
			row := table.Row()
			row.Column(x.Host)
			row.Column(x.Status)
			row.Column(x.Description)
			row.Column(fmt.Sprintf("%v", x.Ack))
			row.Column(x.Url)
		})
		if runtime.GOOS == "windows" {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
		} else {
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
		table.Print(os.Stdout)

		// detect changes and send notification
		if viper.GetBool("notify") && prevItems != nil {
			newItems, _ := lo.Difference[Item](items, prevItems)
			for _, item := range newItems {
				log.Debug().Str("type", "new_item").Str("item", fmt.Sprintf("%v", item)).Send()
				err := beeep.Notify(fmt.Sprintf("%s - %s", item.Status, item.Host), item.Description, "assets/information.png")
				if err != nil {
					panic(err)
				}
			}
		}

		time.Sleep(time.Duration(viper.GetInt("refresh")) * time.Second)
	}
}

func main() {
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
