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

func TestValueFromSelectors(t *testing.T) {
	testRow := `{"level":"info","msg":"hello world","time":"2020-01-01T00:00:00Z"}`

	// single selector "." should yield the whole row
	value := valueGetterFromSelectors([]string{"."}, "", nil)(testRow)
	if value != testRow {
		t.Errorf("Expected value to be '%s', got '%s'", testRow, value)
	}

	// single selector "json(level)" should yield "info"
	value = valueGetterFromSelectors([]string{"json(level)"}, "", nil)(testRow)
	if value != "info" {
		t.Errorf("Expected value to be 'info', got '%s'", value)
	}

	// single selector "json(non-existing-key)" should yield ""
	value = valueGetterFromSelectors([]string{"json(non-existing-key)"}, "", nil)(testRow)
	if value != "" {
		t.Errorf("Expected value to be '', got '%s'", value)
	}

	// two selectors "json(msg)" and "json(level)" should yield "hello world" (first match)
	value = valueGetterFromSelectors([]string{"json(msg)", "json(level)"}, "", nil)(testRow)
	if value != "hello world" {
		t.Errorf("Expected value to be 'hello world', got '%s'", value)
	}

	// two selectors "json(non-existing-key)" and "json(level)" should yield "info" (first match)
	value = valueGetterFromSelectors(
		[]string{"json(non-existing-key)", "json(level)"},
		"",
		nil,
	)(
		testRow,
	)
	if value != "info" {
		t.Errorf("Expected value to be 'info', got '%s'", value)
	}
}
