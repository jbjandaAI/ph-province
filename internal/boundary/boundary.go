package boundary

import (
	_ "embed"
	"encoding/json"
	"fmt"
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

//go:embed cebu.geojson
var cebuJSON []byte

type geoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates [][][][]float64 `json:"coordinates"`
}

// Cebu returns the embedded Cebu province boundary.
func Cebu() (MultiPolygon, error) {
	var geometry geoJSONGeometry
	if err := json.Unmarshal(cebuJSON, &geometry); err != nil {
		return nil, fmt.Errorf("decode embedded Cebu boundary: %w", err)
	}
	if geometry.Type != "MultiPolygon" {
		return nil, fmt.Errorf("unexpected Cebu geometry type %q", geometry.Type)
	}

	shape := make(MultiPolygon, 0, len(geometry.Coordinates))
	for polygonIndex, rawPolygon := range geometry.Coordinates {
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
