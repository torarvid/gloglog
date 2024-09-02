package config

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pelletier/go-toml/v2"
)

type ConfigError string

func (e ConfigError) Error() string {
	return string(e)
}

var (
	ErrNoAttrWithName  = ConfigError("no attribute with name")
	ErrInvalidFilterOp = ConfigError("invalid filter op")
)

type Config struct {
	SavedViews []*LogView
	activeView *LogView
}

// LogView represents a named view of a log source.
//
// A SourceId can be "file", [or in the future...] a database, a remote service like Logstash or
// Loki etc.
//
// Some of these sources require additional options to be specified; Options can be used for that.
//
// Attributes are the columns that are shown in the view.
//
// Filters are used to filter which rows are shown.
type LogView struct {
	Name     string
	SourceId string
	Options  map[string]string
	Attrs    []Attribute
	Filters  []Filter
}

func (lv LogView) GetAttributeWithName(name string) (*Attribute, error) {
	for _, attr := range lv.Attrs {
		if attr.Name == name {
			return &attr, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrNoAttrWithName, name)
}

func (lv LogView) GetRows() []string {
	switch lv.SourceId {
	case "file":
		return fromFile(lv)
	default:
		panic("Unknown source id: " + lv.SourceId)
	}
}

// global config 🤘
var TheConfig *Config

func Load() *Config {
	reader, err := os.Open(getFilePath())
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	return LoadFrom(reader)
}

func LoadFrom(reader io.Reader) *Config {
	configBytes, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	var config Config
	err = toml.Unmarshal(configBytes, &config)
	if err != nil {
		panic(err)
	}
	TheConfig = &config
	return &config
}

func (c Config) Save() {
	writer, err := os.Create(getFilePath())
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	c.SaveTo(writer)
}

func (c Config) SaveTo(writer io.Writer) {
	configBytes, err := toml.Marshal(c)
	if err != nil {
		panic(err)
	}

	_, err = writer.Write(configBytes)
	if err != nil {
		panic(err)
	}
}

func (c *Config) GetActiveView() *LogView {
	if c.activeView == nil {
		c.SetActiveView(c.SavedViews[0])
	}
	return c.activeView
}

func (c *Config) SetActiveView(view *LogView) {
	c.activeView = view
}

func getFilePath() string {
	folder := os.Getenv("XDG_CONFIG_HOME")
	if folder == "" {
		var err error
		folder, err = os.UserConfigDir()
		if err != nil {
			panic(err)
		}
	}
	return path.Join(folder, "gloglog", "config.toml")
}

type Attribute struct {
	Name      string
	Width     int
	Selectors []string
	Type      string
	Format    *string
}

type FilterOp string

const (
	Equal              FilterOp = "=="
	NotEqual           FilterOp = "!="
	RegexEqual         FilterOp = "=~"
	RegexNotEqual      FilterOp = "!~"
	Contains           FilterOp = "contains"
	NotContains        FilterOp = "not contains"
	GreaterThan        FilterOp = ">"
	LessThan           FilterOp = "<"
	GreaterThanOrEqual FilterOp = ">="
	LessThanOrEqual    FilterOp = "<="
)

func ParseFilterOp(input string) (FilterOp, error) {
	switch input {
	case "==":
		return Equal, nil
	case "!=":
		return NotEqual, nil
	case "=~":
		return RegexEqual, nil
	case "!~":
		return RegexNotEqual, nil
	case "contains":
		return Contains, nil
	case "not contains":
		return NotContains, nil
	case ">":
		return GreaterThan, nil
	case "<":
		return LessThan, nil
	case ">=":
		return GreaterThanOrEqual, nil
	case "<=":
		return LessThanOrEqual, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidFilterOp, input)
	}
}

type Filter struct {
	Term     string
	Operator FilterOp
	Attr     *string
}

func NewFilter(term string) Filter {
	return Filter{Term: term}
}

func (f Filter) WithOp(op FilterOp) Filter {
	f.Operator = op
	return f
}

func (f Filter) WithAttr(attr string) Filter {
	f.Attr = &attr
	return f
}

func (f Filter) String() string {
	attr := "[anything]"
	if f.Attr != nil {
		attr = *f.Attr
	}
	op := "contains"
	if f.Operator != "" {
		op = string(f.Operator)
	}
	return fmt.Sprintf("'%s %s %s'", attr, op, f.Term)
}
