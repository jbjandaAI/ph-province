package boundary

import "testing"

func TestCebu(t *testing.T) {
	shape, err := Cebu()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(shape), 52; got != want {
		t.Fatalf("polygon count = %d, want %d", got, want)
	}
	for index, polygon := range shape {
		if len(polygon) == 0 || len(polygon[0]) < 4 {
			t.Fatalf("polygon %d has no valid outer ring", index)
		}
	}
}
