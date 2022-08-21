package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cavaliercoder/go-zabbix"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var triggerSeverity = map[string]int{
	"unknown":     zabbix.TriggerSeverityNotClassified,
	"information": zabbix.TriggerSeverityInformation,
	"warning":     zabbix.TriggerSeverityWarning,
	"average":     zabbix.TriggerSeverityAverage,
	"high":        zabbix.TriggerSeverityHigh,
	"critical":    zabbix.TriggerSeverityDisaster,
}

var hostAvalability = map[string]int{
	"UNKNOWN":     zabbix.HostInterfaceAvailabilityUnknown,
	"AVAILABLE":   zabbix.HostInterfaceAvailabilityAvailable,
	"UNAVAILABLE": zabbix.HostInterfaceAvailabilityUnavailable,
}

type zabbixmonItem struct {
	Host        string `json:"host"`
	Status      string `json:"status"`
	Description string `json:"desc"`
	Time        string `json:"time"`
	Ack         bool   `json:"ack"`
	Url         string `json:"url"`
}

func getSession(server string, username string, password string) *zabbix.Session {
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

func getTriggers(zapi *zabbix.Session, minSeverity string) (triggerItemsUnack []zabbixmonItem, triggerItemsAck []zabbixmonItem) {
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
	triggers = lo.Filter(triggers, func(x zabbix.Trigger, _ int) bool { return x.LastEvent.Value == 1 })

	// tranform triggers into structured items
	severity := lo.Invert(triggerSeverity)
	triggerItemsAll := lo.Map(triggers, func(x zabbix.Trigger, _ int) zabbixmonItem {
		return zabbixmonItem{
			Host:        x.Hosts[0].Hostname,
			Status:      strings.ToUpper(severity[x.Severity]),
			Description: x.Description,
			Time:        x.LastEvent.Timestamp.Format("2006-01-02 15:04"),
			Ack:         x.LastEvent.Acknowledged,
			Url:         fmt.Sprintf("%s/tr_events.php?triggerid=%s&eventid=%s", server, x.TriggerID, x.LastEvent.EventID),
		}
	})
	log.Debug().Str("type", "triggers").Str("scope", "all").Str("items", fmt.Sprintf("%#v", triggerItemsAll)).Send()

	// filter unacknowledged items
	triggerItemsUnack = lo.Filter(triggerItemsAll, func(x zabbixmonItem, _ int) bool { return !x.Ack })
	log.Debug().Str("type", "triggers").Str("scope", "unack").Str("items", fmt.Sprintf("%#v", triggerItemsUnack)).Send()

	// filter acknowledged items
	triggerItemsAck = lo.Filter(triggerItemsAll, func(x zabbixmonItem, _ int) bool { return x.Ack })
	log.Debug().Str("type", "triggers").Str("scope", "ack").Str("items", fmt.Sprintf("%#v", triggerItemsAck)).Send()

	return triggerItemsUnack, triggerItemsAck
}

func getHosts(zapi *zabbix.Session) (hostItemsUnavailable []zabbixmonItem, hostItemsUnknown []zabbixmonItem) {
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
	availability := lo.Invert(hostAvalability)
	hostItemsAll := lo.Map(hosts, func(x zabbix.Host, _ int) zabbixmonItem {
		return zabbixmonItem{
			Host:        x.Hostname,
			Status:      availability[x.Available],
			Description: fmt.Sprintf("Host in %s state", availability[x.Available]),
			Time:        "-",
			Ack:         false,
			Url:         fmt.Sprintf("%s/hostinventories.php?hostid=%s", server, x.HostID),
		}
	})
	log.Debug().Str("type", "hosts").Str("scope", "all").Str("items", fmt.Sprintf("%#v", hostItemsAll)).Send()

	// filter unavailable items
	hostItemsUnavailable = lo.Filter(hostItemsAll, func(x zabbixmonItem, _ int) bool { return x.Status == "UNAVAILABLE" })
	log.Debug().Str("type", "hosts").Str("scope", "unavailable").Str("items", fmt.Sprintf("%#v", hostItemsUnavailable)).Send()

	// filter unknown items
	hostItemsUnknown = lo.Filter(hostItemsAll, func(x zabbixmonItem, _ int) bool { return x.Status == "UNKNOWN" })
	log.Debug().Str("type", "hosts").Str("scope", "unknown").Str("items", fmt.Sprintf("%#v", hostItemsUnknown)).Send()

	return hostItemsUnavailable, hostItemsUnknown
}

func getItems(zapi *zabbix.Session, itemTypes []string, minSeverity string, grep string) (items []zabbixmonItem) {
	// get triggers
	triggerItemsUnack, triggerItemsAck := getTriggers(zapi, minSeverity)
	if present := lo.Contains(itemTypes, "unack"); present {
		items = append(items, triggerItemsUnack...)
	}
	if present := lo.Contains(itemTypes, "ack"); present {
		items = append(items, triggerItemsAck...)
	}

	// get hosts
	hostItemsUnavailable, hostItemsUnknown := getHosts(zapi)
	if present := lo.Contains(itemTypes, "down"); present {
		items = append(items, hostItemsUnavailable...)
	}
	if present := lo.Contains(itemTypes, "unknown"); present {
		items = append(items, hostItemsUnknown...)
	}

	// filter items on hostnames
	if grep != "" {
		hostRegexp := regexp.MustCompile(grep)
		items = lo.Filter(items, func(x zabbixmonItem, _ int) bool {
			return hostRegexp.MatchString(x.Host)
		})
	}

	return items
}
