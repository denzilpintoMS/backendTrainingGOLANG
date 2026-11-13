package main

import "testing"

func TestGradeCalc(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		wantGrade rune
	}{
		{"minimum score", 0, 'F'},
		{"highest F", 54, 'F'},
		{"lowest D", 55, 'D'},
		{"highest D", 64, 'D'},
		{"lowest C", 65, 'C'},
		{"highest C", 74, 'C'},
		{"lowest B", 75, 'B'},
		{"highest B", 84, 'B'},
		{"lowest A", 85, 'A'},
		{"maximum score", 100, 'A'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGrade := gradeCalc(tt.input)
			if gotGrade != tt.wantGrade {
				t.Errorf("got grade %c, want grade %c", gotGrade, tt.wantGrade)
			}
		})
	}
}
