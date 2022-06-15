package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/markkurossi/tabulate"
	"github.com/samber/lo"
)

func TestBuildTableHeaders(t *testing.T) {
	tab := buildTable([]zabbixmonItem{})
	headers_strings := []string{"Host", "Status", "Description", "Ack", "URL"}
	headers := lo.Map[*tabulate.Column, string](tab.Headers, func(x *tabulate.Column, _ int) string {
		return x.Data.String()
	})
	if union := lo.Union[string](headers_strings, headers); len(union) != len(headers_strings) {
		t.Errorf("got: %v, want: %v", headers, headers_strings)
	}
}

func TestBuildTableRows(t *testing.T) {
	item := zabbixmonItem{"myhost", "mystatus", "mydesc", true, "myurl"}
	item_strings := []string{"myhost", "mystatus", "mydesc", "true", "myurl"}
	tab := buildTable([]zabbixmonItem{item})
	row := lo.Map[*tabulate.Column, string](tab.Rows[0].Columns, func(x *tabulate.Column, _ int) string {
		return x.Data.String()
	})
	if union := lo.Union[string](item_strings, row); len(union) != len(item_strings) {
		t.Errorf("got: %v, want: %v", row, item_strings)
	}
}

func TestInitConfigNotFound(t *testing.T) {
	if os.Getenv("CONF_NOT_FOUND") == "1" {
		initConfig()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInitConfigNotFound")
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	cmd.Env = append(os.Environ(), "CONF_NOT_FOUND=1")
	if err := cmd.Run(); err == nil {
		t.Errorf("got: %v, want exit status 1", err)
	}
}
