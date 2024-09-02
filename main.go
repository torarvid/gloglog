package main

import (
	"log/slog"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torarvid/gloglog/config"
)

var (
	appStartTime time.Time = time.Now()
	firstDraw    bool
)

func main() {
	initLogger()
	config := config.Load()
	cfgLoadTime := time.Since(appStartTime)
	slog.Info("Config loaded in", "time", cfgLoadTime)
	view := config.GetActiveView()

	m := newModel(*view)
	modelInitTime := time.Since(appStartTime) - cfgLoadTime
	slog.Info("Model initialized in", "time", modelInitTime)
	program := tea.NewProgram(m, tea.WithAltScreen())
	_, err := program.Run()
	if err != nil {
		slog.Info("Error running program:", "error", err)
		os.Exit(1)
	}
}

func initLogger() {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	handler := slog.NewTextHandler(f, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
