package main

import (
	"testing"
)

func TestEvenOrOdd(t *testing.T) {

	tests := []struct {
		name           string
		input          []int
		wantEvenCounts int
		wantOddCounts  int
	}{
		{"case 1 happy path", []int{1, 2, 3, 4, 5, 6}, 3, 3},
		{"case 2 0 values", []int{0, 0, 0, 0}, 4, 0},
		{"case 3 negative values", []int{-1, -2, 3, 4}, 2, 2},
		{"case 4 empty values", []int{}, 0, 0},
		{"case 5 single value", []int{5}, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvnCounts, gotOddCounts := EvenOrOdd(tt.input)
			if gotEvnCounts != tt.wantEvenCounts {
				t.Errorf("got even counts %v want even counts %v", gotEvnCounts, tt.wantEvenCounts)
			}
			if gotOddCounts != tt.wantOddCounts {
				t.Errorf("got odd counts %v, want odd counts %v", gotOddCounts, tt.wantOddCounts)
			}
		})
	}
}
