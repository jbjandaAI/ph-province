package terminal

import (
	"os"
	"testing"
)

func TestEnvironmentFallback(t *testing.T) {
	t.Setenv("COLUMNS", "101")
	t.Setenv("LINES", "37")
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	columns, rows := Size(file)
	if columns != 101 || rows != 37 {
		t.Fatalf("Size = %dx%d, want 101x37", columns, rows)
	}
}

func TestInvalidEnvironmentFallback(t *testing.T) {
	t.Setenv("COLUMNS", "invalid")
	t.Setenv("LINES", "0")
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	columns, rows := Size(file)
	if columns != defaultColumns || rows != defaultRows {
		t.Fatalf("Size = %dx%d, want %dx%d", columns, rows, defaultColumns, defaultRows)
	}
}
