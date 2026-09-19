package boundary

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestCatalogContains82UniqueProvinces(t *testing.T) {
	names, err := Names()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(names), 82; got != want {
		t.Fatalf("province count = %d, want %d", got, want)
	}
	if !slices.IsSortedFunc(names, func(first, second string) int {
		return strings.Compare(strings.ToLower(first), strings.ToLower(second))
	}) {
		t.Fatal("province names are not sorted")
	}
	codes := make(map[string]struct{}, len(names))
	for _, name := range names {
		province, err := Resolve(name)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", name, err)
		}
		if len(province.Shape) == 0 {
			t.Fatalf("%s has no polygons", name)
		}
		if _, duplicate := codes[province.Code]; duplicate {
			t.Fatalf("duplicate PSGC code %s", province.Code)
		}
		codes[province.Code] = struct{}{}
	}
}

func TestResolveNormalizationAndAliases(t *testing.T) {
	tests := map[string]string{
		"  CeBu  ":                  "Cebu",
		"province of tawi tawi":     "Tawi-Tawi",
		"Compostela Valley":         "Davao de Oro",
		"north cotabato":            "Cotabato",
		"Western Samar":             "Samar",
		"Dinagat Island":            "Dinagat Islands",
		"maguindanao-del-sur":       "Maguindanao del Sur",
		"Agusan,    del...Norte!!!": "Agusan del Norte",
	}
	for input, want := range tests {
		province, err := Resolve(input)
		if err != nil {
			t.Errorf("Resolve(%q): %v", input, err)
			continue
		}
		if province.Name != want {
			t.Errorf("Resolve(%q) = %q, want %q", input, province.Name, want)
		}
	}
}

func TestResolveAmbiguousMaguindanao(t *testing.T) {
	_, err := Resolve("Maguindanao")
	var matchError *MatchError
	if !errors.As(err, &matchError) || !matchError.Ambiguous {
		t.Fatalf("error = %v, want ambiguous MatchError", err)
	}
	want := []string{"Maguindanao del Norte", "Maguindanao del Sur"}
	if !slices.Equal(matchError.Suggestions, want) {
		t.Fatalf("suggestions = %v, want %v", matchError.Suggestions, want)
	}
}

func TestResolveTypoSuggestion(t *testing.T) {
	_, err := Resolve("Boholl")
	var matchError *MatchError
	if !errors.As(err, &matchError) {
		t.Fatalf("error = %v, want MatchError", err)
	}
	if len(matchError.Suggestions) == 0 || matchError.Suggestions[0] != "Bohol" {
		t.Fatalf("suggestions = %v, want Bohol first", matchError.Suggestions)
	}
}

func TestDecodePolygonAndMultiPolygon(t *testing.T) {
	polygon := jsonBytes(`[[[0,0],[1,0],[1,1],[0,0]]]`)
	if shape, err := decodeGeometry("Polygon", polygon); err != nil || len(shape) != 1 {
		t.Fatalf("Polygon shape=%v err=%v", shape, err)
	}
	multiPolygon := jsonBytes(`[[[[0,0],[1,0],[1,1],[0,0]]]]`)
	if shape, err := decodeGeometry("MultiPolygon", multiPolygon); err != nil || len(shape) != 1 {
		t.Fatalf("MultiPolygon shape=%v err=%v", shape, err)
	}
}

func jsonBytes(value string) []byte { return []byte(value) }
