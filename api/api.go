package api

import (
	"fmt"
	"strings"

	"github.com/cavaliercoder/go-zabbix"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

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

func GetSession(server string, username string, password string) *zabbix.Session {
	// authenticate to zabbix server
	zapi, err := zabbix.NewSession(
		server+"/api_jsonrpc.php",
		username,
		password,
	)
	if err != nil {
		log.Error().Err(err).Send()
	}

	return zapi
}

func GetTriggers(zapi *zabbix.Session, minSeverity string) (triggerItemsUnack []Item, triggerItemsAck []Item) {
	server := zapi.URL[:strings.LastIndex(zapi.URL, "/")]

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
		MinSeverity:       triggerSeverity[minSeverity],
	}
	triggers, err := zapi.GetTriggers(triggerParams)
	if err != nil {
		log.Error().Err(err).Send()
	}
	log.Debug().Str("type", "triggers_raw").Str("scope", "all").Str("triggers", fmt.Sprintf("%#v", triggers)).Send()

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
			Url:         fmt.Sprintf("%s/tr_events.php?triggerid=%s&eventid=%s", server, x.TriggerID, x.LastEvent.EventID),
		}
	})
	log.Debug().Str("type", "triggers").Str("scope", "all").Str("items", fmt.Sprintf("%#v", triggerItemsAll)).Send()

	// filter unacknowledged items
	triggerItemsUnack = lo.Filter[Item](triggerItemsAll, func(x Item, _ int) bool { return !x.Ack })
	log.Debug().Str("type", "triggers").Str("scope", "unack").Str("items", fmt.Sprintf("%#v", triggerItemsUnack)).Send()

	// filter acknowledged items
	triggerItemsAck = lo.Filter[Item](triggerItemsAll, func(x Item, _ int) bool { return x.Ack })
	log.Debug().Str("type", "triggers").Str("scope", "ack").Str("items", fmt.Sprintf("%#v", triggerItemsAck)).Send()

	return triggerItemsUnack, triggerItemsAck
}

func GetHosts(zapi *zabbix.Session) (hostItemsUnavailable []Item, hostItemsUnknown []Item) {
	server := zapi.URL[:strings.LastIndex(zapi.URL, "/")]

	// query hosts with problems
	hostParams := zabbix.HostGetParams{
		GetParameters: zabbix.GetParameters{
			Filter:       map[string]interface{}{"available": []int{hostAvalability["UNKNOWN"], hostAvalability["UNAVAILABLE"]}},
			OutputFields: []string{"hostid", "host", "available"},
		},
	}
	hosts, err := zapi.GetHosts(hostParams)
	if err != nil {
		log.Error().Err(err).Send()
	}
	log.Debug().Str("type", "hosts_raw").Str("scope", "all").Str("hosts", fmt.Sprintf("%#v", hosts)).Send()

	// tranform hosts into structured items
	availability := lo.Invert[string, int](hostAvalability)
	hostItemsAll := lo.Map[zabbix.Host, Item](hosts, func(x zabbix.Host, _ int) Item {
		return Item{
			Host:        x.Hostname,
			Status:      availability[x.Available],
			Description: fmt.Sprintf("Host in %s state", availability[x.Available]),
			Ack:         false,
			Url:         fmt.Sprintf("%s/hostinventories.php?hostid=%s", server, x.HostID),
		}
	})
	log.Debug().Str("type", "hosts").Str("scope", "all").Str("items", fmt.Sprintf("%#v", hostItemsAll)).Send()

	// filter unavailable items
	hostItemsUnavailable = lo.Filter[Item](hostItemsAll, func(x Item, _ int) bool { return x.Status == "UNAVAILABLE" })
	log.Debug().Str("type", "hosts").Str("scope", "unavailable").Str("items", fmt.Sprintf("%#v", hostItemsUnavailable)).Send()

	// filter unknown items
	hostItemsUnknown = lo.Filter[Item](hostItemsAll, func(x Item, _ int) bool { return x.Status == "UNKNOWN" })
	log.Debug().Str("type", "hosts").Str("scope", "unknown").Str("items", fmt.Sprintf("%#v", hostItemsUnknown)).Send()

	return hostItemsUnavailable, hostItemsUnknown
}

func GetItems(zapi *zabbix.Session, itemTypes []string, minSeverity string) (items []Item) {
	// get triggers
	triggerItemsUnack, triggerItemsAck := GetTriggers(zapi, minSeverity)
	if present := lo.Contains[string](itemTypes, "unack"); present {
		items = append(items, triggerItemsUnack...)
	}
	if present := lo.Contains[string](itemTypes, "ack"); present {
		items = append(items, triggerItemsAck...)
	}

	// get hosts
	hostItemsUnavailable, hostItemsUnknown := GetHosts(zapi)
	if present := lo.Contains[string](itemTypes, "down"); present {
		items = append(items, hostItemsUnavailable...)
	}
	if present := lo.Contains[string](itemTypes, "unknown"); present {
		items = append(items, hostItemsUnknown...)
	}

	return items
}
