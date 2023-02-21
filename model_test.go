package main

import (
	"testing"

	"github.com/torarvid/gloglog/config"
)

func TestFilterSimple(t *testing.T) {
	testRows := []string{
		`{"level":"info","msg":"hello world","time":"2020-01-01T00:00:00Z"}`,
		`{"level":"info","msg":"goodbye world","time":"2020-01-01T00:00:00Z"}`,
	}
	mainModel := model{rows: testRows}

	// no filters should yield all rows
	mainModel.SetFilters([]config.Filter{})
	if len(mainModel.filteredRows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(mainModel.filteredRows))
	}

	// filter for "hello" should yield one row
	mainModel.SetFilters([]config.Filter{{Term: "hello"}})
	if len(mainModel.filteredRows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(mainModel.filteredRows))
	}

	// filter for "hello" and "goodbye" should yield no rows
	mainModel.SetFilters([]config.Filter{{Term: "hello"}, {Term: "goodbye"}})
	if len(mainModel.filteredRows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(mainModel.filteredRows))
	}
}
