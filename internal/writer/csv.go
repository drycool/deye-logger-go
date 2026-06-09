package writer

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drycool/deye-logger-go/internal/registers"
)

// CSVWriter appends metrics snapshots to a CSV file.
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
	names  []string
}

// NewCSV creates and opens a CSV writer.
func NewCSV(path string, includeUnverified bool) (*CSVWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	exists := false
	if _, err := os.Stat(path); err == nil {
		exists = true
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	names := registers.Names(includeUnverified)
	w := csv.NewWriter(f)

	if !exists {
		header := append([]string{"ts"}, names...)
		if err := w.Write(header); err != nil {
			f.Close()
			return nil, err
		}
		w.Flush()
	}

	return &CSVWriter{file: f, writer: w, names: names}, nil
}

// Write appends one snapshot.
func (c *CSVWriter) Write(ts string, values map[string]*float64) {
	row := make([]string, 0, len(c.names)+1)
	row = append(row, ts)
	for _, name := range c.names {
		v, ok := values[name]
		if ok && v != nil {
			row = append(row, fmt.Sprintf("%.2f", *v))
		} else {
			row = append(row, "")
		}
	}
	c.writer.Write(row)
	c.writer.Flush()
}

// Close closes the underlying file.
func (c *CSVWriter) Close() {
	if c.file != nil {
		c.file.Close()
	}
}
