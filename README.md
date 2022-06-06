# ZabbixMon

CLI application to show currents alerts on Zabbix. Like [Nagstamon](https://nagstamon.de) in your terminal, for Zabbix only.

## Installation

Ensure `$GOPATH/bin` is in your `$PATH` and execute:

```
go install github.com/nikaro/zabbixmon@latest
```

From sources:

```
make
sudo make install
```

You can also download pre-compiled binaries in [Releases](https://github.com/nikaro/zabbixmon/releases).

## Usage

```
$ zabbixmon -h
Zabbix Status Monitoring

Usage:
  zabbixmon [flags]

Flags:
  -h, --help                  help for zabbixmon
  -i, --item-types strings    items state types
  -l, --log-level string      logging level
  -m, --min-severity string   minimum trigger severity
  -n, --notify                enable notifications
  -p, --password string       zabbix password
  -r, --refresh int           data refreshing interval
  -s, --server string         zabbix server url
  -u, --username string       zabbix username
```

## Configuration

Copy [config.dist.toml](config.dist.toml) in one of these locations:

* `/etc/zabbixmon/config.toml`
* `$XDG_CONFIG_HOME/zabbixmon/config.toml`
* `$HOME/.config/zabbixmon/config.toml`
* `$HOME/.zabbixmon/config.toml`
* `./config.toml`

## Demo

[![asciicast](https://asciinema.org/a/hc8qbg4UDdbsaSy4wiXEjAY2s.svg)](https://asciinema.org/a/hc8qbg4UDdbsaSy4wiXEjAY2s)
