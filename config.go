package main

type Config struct {
	SavedSearches []SavedSearch
}

type SavedSearch struct {
	Name     string
	SourceId string
	Options  map[string]string
	Columns  []SearchColumn
}

type SearchColumn struct {
	Name   string
	Width  int
	Path   string
	Type   string
	Format *string
}
