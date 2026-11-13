package main

import "testing"

func TestCleanUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"case 1 normal", " Rahul Kumar ", "rahul_kumar"},
		{"case 2 uppercase", "JOHN DOE", "john_doe"},
		{"case 3 empty string", "", ""},
		{"case 4 only spaces", "    ", ""},
		{"case 5 multiple spaces", "John  Doe", "john__doe"},
		{"case 6 single word", "HELLO", "hello"},
		{"case 7 numbers", "USER 123", "user_123"},
		{"case 8 special characters", "John@Doe", "john@doe"},
		{"case 9 already clean", "john_doe", "john_doe"},
		{"case 10 leading spaces", "   Rahul", "rahul"},
		{"case 11 trailing spaces", "Rahul   ", "rahul"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanUserName(tt.input)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
