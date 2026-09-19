package terminal

import (
	"os"
	"strconv"
)

const (
	defaultColumns = 80
	defaultRows    = 24
)

// Size returns the output terminal size, then environment overrides, and
// finally a deterministic 80x24 fallback for redirected output.
func Size(file *os.File) (int, int) {
	if columns, rows, ok := nativeSize(file); ok && columns > 0 && rows > 0 {
		return columns, rows
	}
	columns := positiveEnvironmentInteger("COLUMNS", defaultColumns)
	rows := positiveEnvironmentInteger("LINES", defaultRows)
	return columns, rows
}

func positiveEnvironmentInteger(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
