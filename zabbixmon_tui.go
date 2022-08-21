package main

import (
	"time"

	"github.com/cavaliercoder/go-zabbix"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type model struct {
	zapi      *zabbix.Session
	prevItems []zabbixmonItem
	items     []zabbixmonItem
	refresh   int
	table     table.Model
}

type tickMsg time.Time

var baseStyle = lipgloss.NewStyle()

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "o", "enter":
			if err := openUrl(m.items[m.table.Cursor()].Url); err != nil {
				log.Warn().Str("scope", "opening url").Err(err).Send()
			}

		case "r":
			m.prevItems = append([]zabbixmonItem(nil), m.items...)
			m.items = getItems(m.zapi, config.ItemTypes, config.MinSeverity, config.Grep)
			cursor := m.table.Cursor()
			m.table = updateTable(m.items)
			m.table.SetCursor(cursor)
			m.refresh = config.Refresh
			notify(m.items, m.prevItems)

		case "ctrl+c", "q":
			return m, tea.Quit

		}

	case tickMsg:
		m.refresh -= 1
		if m.refresh <= 0 {
			m.prevItems = append([]zabbixmonItem(nil), m.items...)
			m.items = getItems(m.zapi, config.ItemTypes, config.MinSeverity, config.Grep)
			cursor := m.table.Cursor()
			m.table = updateTable(m.items)
			m.table.SetCursor(cursor)
			m.refresh = config.Refresh
			notify(m.items, m.prevItems)
		}
		return m, tick()

	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View() + "\n")
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// update items table
func updateTable(items []zabbixmonItem) table.Model {
	var maxHostWidth, maxStatusWidth, maxDescWidth, maxTimeWidth int

	rows := []table.Row{}
	for _, item := range items {
		maxHostWidth = lo.Ternary(len([]rune(item.Host)) > maxHostWidth, len([]rune(item.Host)), maxHostWidth)
		maxStatusWidth = lo.Ternary(len([]rune(item.Status)) > maxStatusWidth, len([]rune(item.Status)), maxStatusWidth)
		maxDescWidth = lo.Ternary(len([]rune(item.Description)) > maxDescWidth, len([]rune(item.Description)), maxDescWidth)
		maxTimeWidth = lo.Ternary(len([]rune(item.Time)) > maxTimeWidth, len([]rune(item.Time)), maxTimeWidth)
		rows = append(rows, []string{item.Host, item.Status, item.Description, item.Time, lo.Ternary(item.Ack, "✓", "✗")})
	}

	columns := []table.Column{
		{Title: "Host", Width: maxHostWidth},
		{Title: "Status", Width: maxStatusWidth},
		{Title: "Description", Width: maxDescWidth},
		{Title: "Time", Width: maxTimeWidth},
		{Title: "Ack", Width: 3},
	}

	return table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
	)
}

func initModel() model {
	// zabbix auth
	zapi := getSession(config.Server, config.Username, config.Password)

	// fetch items
	items := getItems(zapi, config.ItemTypes, config.MinSeverity, config.Grep)

	// build table
	t := updateTable(items)
	s := table.DefaultStyles()
	t.SetStyles(s)

	return model{
		zapi:      zapi,
		prevItems: []zabbixmonItem{},
		items:     items,
		refresh:   config.Refresh,
		table:     t,
	}
}
