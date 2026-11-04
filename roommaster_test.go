package main

import (
	"os"
	"testing"
)

func TestRoomEnvelope(t *testing.T) {
	s, err := roomSeal(`{"pw":"abc123","code":"RDC-42"}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := roomOpen(s, "senha-errada"); err == nil {
		t.Fatal("abriu com a senha errada")
	}
	mp := os.Getenv("MASTER_PW")
	if mp == "" {
		t.Skip("sem MASTER_PW")
	}
	got, err := roomOpen(s, mp)
	if err != nil || got != `{"pw":"abc123","code":"RDC-42"}` {
		t.Fatalf("não abriu: %v %q", err, got)
	}
}
