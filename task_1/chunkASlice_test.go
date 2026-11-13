package main

import (
	"reflect"
	"testing"
)

func TestChunkSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  [][]int
	}{
		{"last chunk is smaller", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"equal chunks", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"chunk larger than input", []int{1, 2, 3}, 5, [][]int{{1, 2, 3}}},
		{"invalid chunk size", []int{1, 2, 3}, 0, [][]int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChunkSlice(tt.input, tt.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
