package main

import (
	"log"
	"os"

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

func main() {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	config := loadConfig()
	view := config.SavedViews[0]

	m := newModel(view)
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
