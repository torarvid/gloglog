package config

type Config struct {
	SavedViews []LogView
}

type LogView struct {
	Name     string
	SourceId string
	Options  map[string]string
	Attrs    []Attribute
}

type Attribute struct {
	Name      string
	Width     int
	Selectors []string
	Type      string
	Format    *string
}
