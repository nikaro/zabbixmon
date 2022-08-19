package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/gen2brain/beeep"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// dump items as json on stdout
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

// open url in the default web browser
func openUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

// send notification for all items
func notify(items, prevItems []zabbixmonItem) {
	if config.Notify && len(prevItems) > 0 {
		newItems, _ := lo.Difference(items, prevItems)

		for _, item := range newItems {
			log.Debug().Str("type", "new_item").Str("item", fmt.Sprintf("%#v", item)).Send()
			if err := beeep.Notify(fmt.Sprintf("%s - %s", item.Status, item.Host), item.Description, "assets/information.png"); err != nil {
				log.Error().Err(err).Str("scope", "sending notification").Send()
				os.Exit(1)
			}
		}
	}
}
