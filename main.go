package main

import (
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pelletier/go-toml/v2"
	"github.com/torarvid/gloglog/config"
)

func loadConfig() config.Config {
	configBytes, err := os.ReadFile("foo.toml")
	if err != nil {
		panic(err)
	}

	var config config.Config
	err = toml.Unmarshal(configBytes, &config)
	if err != nil {
		panic(err)
	}
	return config
}

var appStartTime time.Time = time.Now()
var firstDraw bool

func main() {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	config := loadConfig()
	cfgLoadTime := time.Since(appStartTime)
	log.Println("Config loaded in", cfgLoadTime)
	view := config.SavedViews[0]

	m := newModel(view)
	modelInitTime := time.Since(appStartTime) - cfgLoadTime
	log.Println("Model initialized in", modelInitTime)
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
