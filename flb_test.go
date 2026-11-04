package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlbGuideID(t *testing.T) {
	ok := map[string]string{
		"https://funleaguebuilds.de/guide.html?id=69f7c4587f6adf40e1cb42ab": "69f7c4587f6adf40e1cb42ab",
		"funleaguebuilds.de/guide.html?x=1&id=abc123def":                   "abc123def",
		"  69f7c4587f6adf40e1cb42ab ":                                       "69f7c4587f6adf40e1cb42ab",
	}
	for in, want := range ok {
		if got, err := flbGuideID(in); err != nil || got != want {
			t.Fatalf("%q -> %q %v", in, got, err)
		}
	}
	for _, bad := range []string{"https://evil.com/?id=abcdef12", "https://funleaguebuilds.de/", "oi"} {
		if _, err := flbGuideID(bad); err == nil {
			t.Fatalf("aceitou %q", bad)
		}
	}
}

func TestFlbGuideFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/abcdef123":
			w.Write([]byte(`{"guide":{"id":"abcdef123","champion":"Lux"}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	if s, err := flbGuideFrom(srv.URL+"/api", "abcdef123"); err != nil || s == "" {
		t.Fatal(s, err)
	}
	if _, err := flbGuideFrom(srv.URL+"/api", "zzzzzz999"); err == nil {
		t.Fatal("404 deveria dar erro")
	}
}
