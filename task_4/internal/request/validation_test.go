package request

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRegistrationPasswordLimits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		password string
		valid    bool
	}{
		{"eight unicode characters", strings.Repeat("界", 8), true},
		{"seven unicode characters", strings.Repeat("界", 7), false},
		{"exact bcrypt limit", strings.Repeat("界", 24), true},
		{"exceeds bcrypt bytes", strings.Repeat("界", 25), false},
		{"password whitespace preserved", " secret ", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"name": " Alice ", "email": " ALICE@example.com ", "password": tc.password})
			if err != nil {
				t.Fatal(err)
			}
			var input Register
			err = Decode(body, &input)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
			if input.Name != "Alice" || input.Email != "alice@example.com" || input.Password != tc.password {
				t.Fatal("normalization changed credentials incorrectly")
			}
		})
	}
}

func TestLoginValidation(t *testing.T) {
	for _, body := range []string{
		`{}`, `null`, `[]`, `{"email":"a@example.com","password":""}`,
		`{"email":"invalid","password":"secret"}`,
		`{"email":"Alice <a@example.com>","password":"secret"}`,
		`{"email":"a@example.com","password":"secret","extra":true}`,
		`{"email":"a@example.com","password":"secret"} {}`,
	} {
		var input Login
		if err := Decode([]byte(body), &input); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
	// Login must accept short passwords so credential verification still runs.
	var input Login
	if err := Decode([]byte(`{"email":" A@example.com ","password":"x"}`), &input); err != nil {
		t.Fatal(err)
	}
}
