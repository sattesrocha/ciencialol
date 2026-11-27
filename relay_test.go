package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestRelayCripto(t *testing.T) {
	k := relayDerive(strings.Repeat("ab", 16))
	if !strings.HasPrefix(k.Down, "clold-") || !strings.HasPrefix(k.Up, "clolu-") || k.Down == k.Up {
		t.Fatal(k.Down, k.Up)
	}
	if strings.Contains(k.Down+k.Up, "abab") {
		t.Fatal("a chave não pode aparecer no nome do tópico")
	}
	msg := k.seal(gz([]byte(`{"t":"state"}`)), "pc")
	if _, err := k.open(msg, "cel"); err == nil {
		t.Fatal("mensagem do PC não pode passar como se fosse do celular")
	}
	if _, err := relayDerive(strings.Repeat("cd", 16)).open(msg, "pc"); err == nil {
		t.Fatal("outra chave não pode abrir")
	}
	p, err := k.open(msg, "pc")
	if err != nil {
		t.Fatal(err)
	}
	zr, _ := gzip.NewReader(bytes.NewReader(p))
	b, _ := io.ReadAll(zr)
	var o map[string]string
	if json.Unmarshal(b, &o) != nil || o["t"] != "state" {
		t.Fatalf("%s", b)
	}
}
