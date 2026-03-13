package handler

import (
	"testing"
)

func TestValidToken(t *testing.T) {
	tests := []struct {
		name     string
		envToken string
		provided string
		want     bool
	}{
		{
			name:     "correct token",
			envToken: "supersecret123",
			provided: "supersecret123",
			want:     true,
		},
		{
			name:     "wrong token",
			envToken: "supersecret123",
			provided: "wrongtoken",
			want:     false,
		},
		{
			name:     "empty provided token",
			envToken: "supersecret123",
			provided: "",
			want:     false,
		},
		{
			name:     "empty env token — fail closed",
			envToken: "",
			provided: "supersecret123",
			want:     false,
		},
		{
			name:     "both empty — fail closed",
			envToken: "",
			provided: "",
			want:     false,
		},
		{
			name:     "prefix match does not pass",
			envToken: "supersecret123",
			provided: "supersecret",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ADMIN_TOKEN", tc.envToken)
			got := validToken(tc.provided)
			if got != tc.want {
				t.Errorf("validToken(%q) with env=%q = %v, want %v",
					tc.provided, tc.envToken, got, tc.want)
			}
		})
	}
}

func TestValidDate(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"2026-01-15", true},
		{"2026-12-31", true},
		{"2026-1-5", false},
		{"01-15-2026", false},
		{"not-a-date", false},
		{"2026-01-15; DROP TABLE incidents", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := validDate.MatchString(tc.input)
			if got != tc.want {
				t.Errorf("validDate.MatchString(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
