package boundary

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// Point is a longitude/latitude coordinate before projection and a Cartesian
// coordinate afterwards.
type Point struct {
	X float64
	Y float64
}

// Polygon contains an outer ring followed by zero or more hole rings.
type Polygon [][]Point

// MultiPolygon is a collection of disconnected polygons.
type MultiPolygon []Polygon

// Province is a canonical Philippine province and its geographic boundary.
type Province struct {
	Name  string
	Code  string
	Shape MultiPolygon
}

// MatchError reports an unknown or ambiguous province name.
type MatchError struct {
	Query       string
	Suggestions []string
	Ambiguous   bool
}

func (err *MatchError) Error() string {
	if err.Ambiguous {
		return fmt.Sprintf("ambiguous province %q; choose: %s", err.Query, strings.Join(err.Suggestions, ", "))
	}
	if len(err.Suggestions) > 0 {
		return fmt.Sprintf("unknown province %q; did you mean: %s?", err.Query, strings.Join(err.Suggestions, ", "))
	}
	return fmt.Sprintf("unknown province %q; use --list to see supported provinces", err.Query)
}

//go:embed provinces.geojson
var provincesJSON []byte

type featureCollection struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

type feature struct {
	Properties struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"properties"`
	Geometry struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	} `json:"geometry"`
}

type catalog struct {
	byName map[string]Province
	names  []string
}

var (
	catalogOnce sync.Once
	loaded      catalog
	loadError   error
)

var aliases = map[string]string{
	"compostela valley": "Davao de Oro",
	"north cotabato":    "Cotabato",
	"western samar":     "Samar",
	"dinagat island":    "Dinagat Islands",
}

// Resolve finds a province using normalized official names and supported
// historical aliases.
func Resolve(query string) (Province, error) {
	data, err := getCatalog()
	if err != nil {
		return Province{}, err
	}
	key := normalize(query)
	if key == "maguindanao" {
		return Province{}, &MatchError{
			Query: query, Ambiguous: true,
			Suggestions: []string{"Maguindanao del Norte", "Maguindanao del Sur"},
		}
	}
	if canonical, ok := aliases[key]; ok {
		key = normalize(canonical)
	}
	if province, ok := data.byName[key]; ok {
		return province, nil
	}
	return Province{}, &MatchError{Query: query, Suggestions: suggestions(data, key)}
}

// Names returns an alphabetical copy of all supported canonical names.
func Names() ([]string, error) {
	data, err := getCatalog()
	if err != nil {
		return nil, err
	}
	return append([]string(nil), data.names...), nil
}

func getCatalog() (*catalog, error) {
	catalogOnce.Do(func() {
		loaded, loadError = decodeCatalog(provincesJSON)
	})
	if loadError != nil {
		return nil, loadError
	}
	return &loaded, nil
}

func decodeCatalog(data []byte) (catalog, error) {
	var collection featureCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return catalog{}, fmt.Errorf("decode embedded province boundaries: %w", err)
	}
	if collection.Type != "FeatureCollection" {
		return catalog{}, fmt.Errorf("unexpected boundary collection type %q", collection.Type)
	}

	result := catalog{byName: make(map[string]Province, len(collection.Features))}
	codes := make(map[string]struct{}, len(collection.Features))
	for index, item := range collection.Features {
		if item.Properties.Name == "" || item.Properties.Code == "" {
			return catalog{}, fmt.Errorf("province feature %d is missing its name or PSGC code", index)
		}
		key := normalize(item.Properties.Name)
		if _, exists := result.byName[key]; exists {
			return catalog{}, fmt.Errorf("duplicate province name %q", item.Properties.Name)
		}
		if _, exists := codes[item.Properties.Code]; exists {
			return catalog{}, fmt.Errorf("duplicate province PSGC code %q", item.Properties.Code)
		}
		shape, err := decodeGeometry(item.Geometry.Type, item.Geometry.Coordinates)
		if err != nil {
			return catalog{}, fmt.Errorf("decode %s boundary: %w", item.Properties.Name, err)
		}
		result.byName[key] = Province{Name: item.Properties.Name, Code: item.Properties.Code, Shape: shape}
		result.names = append(result.names, item.Properties.Name)
		codes[item.Properties.Code] = struct{}{}
	}
	if len(result.names) != 82 {
		return catalog{}, fmt.Errorf("embedded boundary catalog has %d provinces, want 82", len(result.names))
	}
	sort.Slice(result.names, func(first, second int) bool {
		return strings.ToLower(result.names[first]) < strings.ToLower(result.names[second])
	})
	return result, nil
}

func decodeGeometry(geometryType string, data json.RawMessage) (MultiPolygon, error) {
	var raw [][][][]float64
	switch geometryType {
	case "MultiPolygon":
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
	case "Polygon":
		var polygon [][][]float64
		if err := json.Unmarshal(data, &polygon); err != nil {
			return nil, err
		}
		raw = [][][][]float64{polygon}
	default:
		return nil, fmt.Errorf("unsupported geometry type %q", geometryType)
	}

	shape := make(MultiPolygon, 0, len(raw))
	for polygonIndex, rawPolygon := range raw {
		polygon := make(Polygon, 0, len(rawPolygon))
		for ringIndex, rawRing := range rawPolygon {
			if len(rawRing) < 4 {
				return nil, fmt.Errorf("polygon %d ring %d has fewer than four points", polygonIndex, ringIndex)
			}
			ring := make([]Point, 0, len(rawRing))
			for _, coordinate := range rawRing {
				if len(coordinate) < 2 {
					return nil, fmt.Errorf("polygon %d ring %d contains an invalid coordinate", polygonIndex, ringIndex)
				}
				ring = append(ring, Point{X: coordinate[0], Y: coordinate[1]})
			}
			polygon = append(polygon, ring)
		}
		shape = append(shape, polygon)
	}
	return shape, nil
}

func normalize(value string) string {
	var output strings.Builder
	separator := false
	for _, character := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if separator && output.Len() > 0 {
				output.WriteByte(' ')
			}
			output.WriteRune(character)
			separator = false
		} else {
			separator = true
		}
	}
	result := output.String()
	return strings.TrimSpace(strings.TrimPrefix(result, "province of "))
}

type scoredName struct {
	name     string
	distance int
}

func suggestions(data *catalog, query string) []string {
	if query == "" {
		return nil
	}
	threshold := max(2, len([]rune(query))/3)
	scores := make([]scoredName, 0, len(data.names))
	for _, name := range data.names {
		distance := editDistance(query, normalize(name))
		if distance <= threshold {
			scores = append(scores, scoredName{name: name, distance: distance})
		}
	}
	sort.Slice(scores, func(first, second int) bool {
		if scores[first].distance == scores[second].distance {
			return scores[first].name < scores[second].name
		}
		return scores[first].distance < scores[second].distance
	})
	if len(scores) > 3 {
		scores = scores[:3]
	}
	result := make([]string, len(scores))
	for index, score := range scores {
		result[index] = score.name
	}
	return result
}

func editDistance(first, second string) int {
	a, b := []rune(first), []rune(second)
	previous := make([]int, len(b)+1)
	for index := range previous {
		previous[index] = index
	}
	for firstIndex, firstRune := range a {
		current := make([]int, len(b)+1)
		current[0] = firstIndex + 1
		for secondIndex, secondRune := range b {
			cost := 0
			if firstRune != secondRune {
				cost = 1
			}
			current[secondIndex+1] = min(
				current[secondIndex]+1,
				previous[secondIndex+1]+1,
				previous[secondIndex]+cost,
			)
		}
		previous = current
	}
	return previous[len(b)]
}
