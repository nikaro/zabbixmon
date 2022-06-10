package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/markkurossi/tabulate"
	"github.com/nikaro/zabbixmon/api"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

func TestSetLogLevelPanic(t *testing.T) {
	setLogLevel("panic")
	if zerolog.GlobalLevel() != zerolog.PanicLevel {
		t.Errorf("got: %v, want: %v", zerolog.GlobalLevel(), zerolog.PanicLevel)
	}
}

func TestSetLogLevelWarn(t *testing.T) {
	setLogLevel("warn")
	if zerolog.GlobalLevel() != zerolog.WarnLevel {
		t.Errorf("got: %v, want: %v", zerolog.GlobalLevel(), zerolog.WarnLevel)
	}
}

func TestSetLogLevelNotFound(t *testing.T) {
	if os.Getenv("LOG_NOT_FOUND") == "1" {
		setLogLevel("not_found")
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestSetLogLevelNotFound")
	cmd.Env = append(os.Environ(), "LOG_NOT_FOUND=1")
	if err := cmd.Run(); err == nil {
		t.Errorf("got: %v, want exit status 1", err)
	}
}

func TestBuildTableHeaders(t *testing.T) {
	tab := buildTable([]api.Item{})
	headers_strings := []string{"Host", "Status", "Description", "Ack", "URL"}
	headers := lo.Map[*tabulate.Column, string](tab.Headers, func(x *tabulate.Column, _ int) string {
		return x.Data.String()
	})
	if union := lo.Union[string](headers_strings, headers); len(union) != len(headers_strings) {
		t.Errorf("got: %v, want: %v", headers, headers_strings)
	}
}

func TestBuildTableRows(t *testing.T) {
	item := api.Item{"myhost", "mystatus", "mydesc", true, "myurl"}
	item_strings := []string{"myhost", "mystatus", "mydesc", "true", "myurl"}
	tab := buildTable([]api.Item{item})
	row := lo.Map[*tabulate.Column, string](tab.Rows[0].Columns, func(x *tabulate.Column, _ int) string {
		return x.Data.String()
	})
	if union := lo.Union[string](item_strings, row); len(union) != len(item_strings) {
		t.Errorf("got: %v, want: %v", row, item_strings)
	}
}
