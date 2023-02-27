package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func fromFile(logView LogView) []string {
	filename, exists := logView.Options["filename"]
	if !exists {
		panic("filename not found")
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Fprintf(os.Stderr, "Scanning file '%s'", filename)
	rows := make([]string, 0)
	for scanner.Scan() {
		rows = append(rows, scanner.Text())
	}
	return rows
}
