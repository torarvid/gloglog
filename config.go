package main

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
	Name     string
	Width    int
	Selector string
	Type     string
	Format   *string
}
