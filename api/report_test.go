package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSha256hex(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Known SHA-256 vectors
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"192.168.1.1", "c5eb5a4cc76a5cdb16e79864b9ccd26c3553f0c396d0a21bafb7be71c1efcd8c"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := sha256hex(tc.input)
			if got != tc.want {
				t.Errorf("sha256hex(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name   string
		xff    string
		remote string
		wantIP string
	}{
		{
			name:   "single IP in X-Forwarded-For",
			xff:    "1.2.3.4",
			remote: "10.0.0.1:9000",
			wantIP: "1.2.3.4",
		},
		{
			name:   "multiple IPs — picks rightmost (real client on Vercel)",
			xff:    "1.1.1.1, 2.2.2.2, 3.3.3.3",
			remote: "10.0.0.1:9000",
			wantIP: "3.3.3.3",
		},
		{
			name:   "no X-Forwarded-For — uses RemoteAddr",
			xff:    "",
			remote: "5.6.7.8:12345",
			wantIP: "5.6.7.8",
		},
		{
			name:   "XFF with spaces around commas",
			xff:    "9.9.9.9 , 8.8.8.8",
			remote: "10.0.0.1:9000",
			wantIP: "8.8.8.8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/report", nil)
			req.RemoteAddr = tc.remote
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			got := clientIP(req)
			if got != tc.wantIP {
				t.Errorf("clientIP() = %q, want %q", got, tc.wantIP)
			}
		})
	}
}
