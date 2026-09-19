package render

import (
	"fmt"
	"math"
	"strings"

	"ph-province/internal/boundary"
)

const (
	minimumColumns = 4
	minimumRows    = 2
)

type bounds struct {
	minX float64
	maxX float64
	minY float64
	maxY float64
}

// Draw converts a geographic multipolygon into centered Unicode half-block
// terminal art that fits within columns and rows.
func Draw(shape boundary.MultiPolygon, columns, rows int) (string, error) {
	if len(shape) == 0 {
		return "", fmt.Errorf("cannot render an empty shape")
	}
	if columns < minimumColumns || rows < minimumRows {
		return "", fmt.Errorf("terminal is too small (minimum %dx%d)", minimumColumns, minimumRows)
	}

	projected, box, err := project(shape)
	if err != nil {
		return "", err
	}

	availableWidth := columns - 2
	availablePixelHeight := rows*2 - 2
	width := box.maxX - box.minX
	height := box.maxY - box.minY
	scale := math.Min(float64(availableWidth)/width, float64(availablePixelHeight)/height)
	rasterWidth := max(1, int(math.Ceil(width*scale)))
	rasterHeight := max(2, int(math.Ceil(height*scale)))
	if rasterHeight%2 != 0 {
		rasterHeight++
	}
	if rasterHeight > rows*2 {
		rasterHeight -= 2
	}

	pixels := make([][]bool, rasterHeight)
	for y := range pixels {
		pixels[y] = make([]bool, rasterWidth)
	}

	for _, polygon := range projected {
		hit := false
		for y := 0; y < rasterHeight; y++ {
			for x := 0; x < rasterWidth; x++ {
				if pixelIntersects(polygon, box, scale, x, y) {
					pixels[y][x] = true
					hit = true
				}
			}
		}

		// Preserve islands that are smaller than one raster pixel.
		if !hit && len(polygon) > 0 && len(polygon[0]) > 0 {
			point := polygon[0][0]
			x := clamp(int((point.X-box.minX)*scale), 0, rasterWidth-1)
			y := clamp(int((point.Y-box.minY)*scale), 0, rasterHeight-1)
			pixels[y][x] = true
		}
	}

	leftPadding := max(0, (columns-rasterWidth)/2)
	var output strings.Builder
	for y := 0; y < rasterHeight; y += 2 {
		var line strings.Builder
		line.WriteString(strings.Repeat(" ", leftPadding))
		for x := 0; x < rasterWidth; x++ {
			upper := pixels[y][x]
			lower := y+1 < rasterHeight && pixels[y+1][x]
			switch {
			case upper && lower:
				line.WriteRune('█')
			case upper:
				line.WriteRune('▀')
			case lower:
				line.WriteRune('▄')
			default:
				line.WriteByte(' ')
			}
		}
		output.WriteString(strings.TrimRight(line.String(), " "))
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func project(shape boundary.MultiPolygon) (boundary.MultiPolygon, bounds, error) {
	box := bounds{
		minX: math.Inf(1), maxX: math.Inf(-1),
		minY: math.Inf(1), maxY: math.Inf(-1),
	}
	minLatitude, maxLatitude := math.Inf(1), math.Inf(-1)
	for _, polygon := range shape {
		for _, ring := range polygon {
			for _, point := range ring {
				minLatitude = math.Min(minLatitude, point.Y)
				maxLatitude = math.Max(maxLatitude, point.Y)
			}
		}
	}
	if math.IsInf(minLatitude, 1) {
		return nil, box, fmt.Errorf("shape has no coordinates")
	}

	longitudeScale := math.Cos((minLatitude + maxLatitude) / 2 * math.Pi / 180)
	projected := make(boundary.MultiPolygon, len(shape))
	for polygonIndex, polygon := range shape {
		projected[polygonIndex] = make(boundary.Polygon, len(polygon))
		for ringIndex, ring := range polygon {
			projected[polygonIndex][ringIndex] = make([]boundary.Point, len(ring))
			for pointIndex, point := range ring {
				projectedPoint := boundary.Point{X: point.X * longitudeScale, Y: -point.Y}
				projected[polygonIndex][ringIndex][pointIndex] = projectedPoint
				box.minX = math.Min(box.minX, projectedPoint.X)
				box.maxX = math.Max(box.maxX, projectedPoint.X)
				box.minY = math.Min(box.minY, projectedPoint.Y)
				box.maxY = math.Max(box.maxY, projectedPoint.Y)
			}
		}
	}
	if box.maxX == box.minX || box.maxY == box.minY {
		return nil, box, fmt.Errorf("shape has zero width or height")
	}
	return projected, box, nil
}

func pixelIntersects(polygon boundary.Polygon, box bounds, scale float64, pixelX, pixelY int) bool {
	for _, offsetY := range []float64{0.25, 0.75} {
		for _, offsetX := range []float64{0.25, 0.75} {
			point := boundary.Point{
				X: box.minX + (float64(pixelX)+offsetX)/scale,
				Y: box.minY + (float64(pixelY)+offsetY)/scale,
			}
			if insidePolygon(point, polygon) {
				return true
			}
		}
	}
	return false
}

func insidePolygon(point boundary.Point, polygon boundary.Polygon) bool {
	if len(polygon) == 0 || !insideRing(point, polygon[0]) {
		return false
	}
	for _, hole := range polygon[1:] {
		if insideRing(point, hole) {
			return false
		}
	}
	return true
}

func insideRing(point boundary.Point, ring []boundary.Point) bool {
	inside := false
	for current, previous := 0, len(ring)-1; current < len(ring); previous, current = current, current+1 {
		a, b := ring[current], ring[previous]
		intersects := (a.Y > point.Y) != (b.Y > point.Y) &&
			point.X < (b.X-a.X)*(point.Y-a.Y)/(b.Y-a.Y)+a.X
		if intersects {
			inside = !inside
		}
	}
	return inside
}

func clamp(value, low, high int) int {
	return min(max(value, low), high)
}
