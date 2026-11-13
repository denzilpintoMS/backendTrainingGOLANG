package main

import "testing"

func TestFindMinMax(t *testing.T) {
	tests := []struct {
		name    string
		input   []int
		wantMin int
		wantMax int
	}{
		{"mixed values", []int{3, 1, 5, 2, 4}, 1, 5},
		{"negative values", []int{-3, -1, -5}, -5, -1},
		{"negative and positive values", []int{-4, 0, 7}, -4, 7},
		{"same values", []int{2, 2, 2}, 2, 2},
		{"single value", []int{5}, 5, 5},
		{"empty input", []int{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := findMinMax(tt.input)
			if gotMin != tt.wantMin {
				t.Errorf("got minimum %v, want minimum %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("got maximum %v, want maximum %v", gotMax, tt.wantMax)
			}
		})
	}
}
