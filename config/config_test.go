package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/torarvid/gloglog/testutil"
)

func TestPaths(t *testing.T) {
	path := getFilePath()
	if strings.HasSuffix(path, "/gloglog/config.toml") == false {
		t.Error("Expected path to end with /gloglog/config.toml, got", path)
	}

	defer TempEnv("XDG_CONFIG_HOME", "/tmp")()
	fmt.Println("XDG_CONFIG_HOME", os.Getenv("XDG_CONFIG_HOME"))
	path = getFilePath()
	if path != "/tmp/gloglog/config.toml" {
		t.Errorf("Expected path to be /tmp/gloglog/config.toml, got %s", path)
	}
}

var validToml = `[[SavedViews]]
Name = 'test'
SourceId = 'file'
Filters = []

[SavedViews.Options]
filename = 'somefile.log'

[[SavedViews.Attrs]]
Name = 'Time'
Width = 12
Selectors = ['json(data) | json(timestamp)']
Type = 'time'
Format = '15:04:05.000'

[[SavedViews.Attrs]]
Name = 'Event'
Width = 30
Selectors = ['json(event)', 'json(data.event)']
Type = 'string'
`

func TestLoadConfig(t *testing.T) {
	invalidToml := `
    [invalid-config-file]
    foo = "bar"
    `
	reader := strings.NewReader(invalidToml)
	config := LoadFrom(reader)
	if len(config.SavedViews) != 0 {
		t.Errorf("Expected no saved views from this invalid config file")
	}

	reader = strings.NewReader(validToml)
	config = LoadFrom(reader)
	AssertEq(t, len(config.SavedViews), 1)
	view := config.SavedViews[0]
	AssertEq(t, view.Name, "test")
	AssertEq(t, view.SourceId, "file")
	AssertEq(t, view.Options["filename"], "somefile.log")
	AssertEq(t, len(view.Attrs), 2)
}

func TestSaveConfig(t *testing.T) {
	reader := strings.NewReader(validToml)
	config := LoadFrom(reader)

	writer := &strings.Builder{}
	config.SaveTo(writer)
	AssertEq(t, validToml, writer.String())

	config.SavedViews[0].Filters = []Filter{{Term: "foo"}}
	writer = &strings.Builder{}
	config.SaveTo(writer)
	expectedToml := strings.Replace(validToml, "Filters = []\n", "", 1)
	expectedToml += "\n[[SavedViews.Filters]]\nTerm = 'foo'\nOperator = ''\n"
	AssertEq(t, expectedToml, writer.String())
}
