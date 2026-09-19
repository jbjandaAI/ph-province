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

func TestRepresentativeProvinceGoldens(t *testing.T) {
	tests := map[string]string{
		"Cebu":                  "cebu.golden",
		"Batanes":               "batanes.golden",
		"Benguet":               "benguet.golden",
		"Maguindanao del Norte": "maguindanao_del_norte.golden",
		"Maguindanao del Sur":   "maguindanao_del_sur.golden",
	}
	for provinceName, filename := range tests {
		t.Run(provinceName, func(t *testing.T) {
			province, err := boundary.Resolve(provinceName)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Draw(province.Shape, 40, 24)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", filename)
			if *updateGolden {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Errorf("rendering differs from %s; run go test ./internal/render -update", path)
			}
			validateOutput(t, got, 40)
		})
	}
}

func TestEveryProvinceRendersAtCommonSizes(t *testing.T) {
	names, err := boundary.Names()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		province, err := boundary.Resolve(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, size := range []struct{ columns, rows int }{{20, 12}, {80, 24}} {
			output, err := Draw(province.Shape, size.columns, size.rows)
			if err != nil {
				t.Errorf("Draw(%s, %dx%d): %v", name, size.columns, size.rows, err)
				continue
			}
			validateOutput(t, output, size.columns)
		}
	}
}

func validateOutput(t *testing.T, output string, columns int) {
	t.Helper()
	if !strings.ContainsAny(output, "▀▄█") {
		t.Error("output does not contain a rendered shape")
	}
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if width := utf8.RuneCountInString(line); width > columns {
			t.Errorf("line width = %d, exceeds %d", width, columns)
		}
		for _, character := range line {
			if character != ' ' && character != '▀' && character != '▄' && character != '█' {
				t.Errorf("unexpected output character %q", character)
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
