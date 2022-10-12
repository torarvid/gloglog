package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pelletier/go-toml/v2"
)

func loadConfig() Config {
	configString, err := ioutil.ReadFile("foo.toml")
	if err != nil {
		panic(err)
	}

	var config Config
	err = toml.Unmarshal([]byte(configString), &config)
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
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
