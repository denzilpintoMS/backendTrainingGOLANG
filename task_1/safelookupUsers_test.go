package main

import "testing"

func TestSafeLookupUser(t *testing.T) {
	users := map[int]string{1: "Alice", 2: "", 3: "Bob"}

	tests := []struct {
		name       string
		userID     int
		wantName   string
		wantExists bool
	}{
		{"existing user", 1, "Alice", true},
		{"missing user", 4, "", false},
		{"user with empty name", 2, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotExists := SafeLookupUser(users, tt.userID)
			if gotName != tt.wantName {
				t.Errorf("got name %q, want name %q", gotName, tt.wantName)
			}
			if gotExists != tt.wantExists {
				t.Errorf("got exists %v, want exists %v", gotExists, tt.wantExists)
			}
		})
	}
}
