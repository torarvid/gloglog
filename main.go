package main

import (
	"log"
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
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	config := config.Load()
	cfgLoadTime := time.Since(appStartTime)
	log.Println("Config loaded in", cfgLoadTime)
	config.SetActiveView(config.SavedViews[0])
	view := *config.GetActiveView()

	m := newModel(view)
	modelInitTime := time.Since(appStartTime) - cfgLoadTime
	log.Println("Model initialized in", modelInitTime)
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
