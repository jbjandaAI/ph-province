package render

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"ph-province/internal/boundary"
)

var updateGolden = flag.Bool("update", false, "update golden rendering files")

func TestCebuGolden(t *testing.T) {
	shape, err := boundary.Cebu()
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []struct {
		columns int
		rows    int
	}{
		{20, 12},
		{40, 24},
		{100, 40},
	} {
		name := filepath.Join("testdata", "cebu_"+itoa(size.columns)+"x"+itoa(size.rows)+".golden")
		got, err := Draw(shape, size.columns, size.rows)
		if err != nil {
			t.Fatal(err)
		}
		if *updateGolden {
			if err := os.WriteFile(name, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if got != string(want) {
			t.Errorf("rendering %dx%d differs from %s; run go test ./internal/render -update", size.columns, size.rows, name)
		}
		for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
			if width := utf8.RuneCountInString(line); width > size.columns {
				t.Errorf("line width = %d, exceeds %d", width, size.columns)
			}
			for _, character := range line {
				if character != ' ' && character != '▀' && character != '▄' && character != '█' {
					t.Errorf("unexpected output character %q", character)
				}
			}
		}
	}
}

func TestPolygonHole(t *testing.T) {
	shape := boundary.MultiPolygon{{
		{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}},
		{{X: 3, Y: 3}, {X: 7, Y: 3}, {X: 7, Y: 7}, {X: 3, Y: 7}, {X: 3, Y: 3}},
	}}
	output, err := Draw(shape, 20, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, " ") {
		t.Fatal("expected blank cells from polygon hole")
	}
}

func TestInvalidInput(t *testing.T) {
	if _, err := Draw(nil, 80, 24); err == nil {
		t.Fatal("expected empty-shape error")
	}
	shape := boundary.MultiPolygon{{{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}}}}
	if _, err := Draw(shape, 3, 1); err == nil {
		t.Fatal("expected terminal-size error")
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
