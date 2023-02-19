package config

import "fmt"

type Config struct {
	SavedViews []LogView
}

type LogView struct {
	Name     string
	SourceId string
	Options  map[string]string
	Attrs    []Attribute
	Filters  []Filter
}

func (lv LogView) GetAttributeWithName(name string) *Attribute {
	for _, attr := range lv.Attrs {
		if attr.Name == name {
			return &attr
		}
	}
	return nil
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
		return "", fmt.Errorf("invalid filter op: %s", input)
	}
}

type Filter struct {
	Term     string
	Operator FilterOp
	Attr     *Attribute
}
