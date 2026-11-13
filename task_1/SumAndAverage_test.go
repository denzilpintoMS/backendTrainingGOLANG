package main

import (
	"testing"
)

func TestSumAndAverage(t *testing.T) {

	tests := []struct {
		name        string
		input       []int
		wantSum     int
		wantAverage int
	}{
		{"case 1 happy path", []int{1, 2, 3}, 6, 2},
		{"case 2 0 values", []int{0, 0, 0, 0}, 0, 0},
		{"case 3 negative values", []int{-1, -2, 3}, 0, 0},
		{"empty input", []int{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSum, gotAverage := SumAndAverage(tt.input)
			if gotSum != tt.wantSum {
				t.Errorf("got sum %v want sum %v", gotSum, tt.wantSum)
			}
			if gotAverage != tt.wantAverage {
				t.Errorf("got average %v, want average %v", gotAverage, tt.wantAverage)
			}
		})
	}
}
