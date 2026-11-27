package main

import (
	"net"
	"testing"
	"time"
)

func TestHandoffNovaJanelaAssumeOCelular(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	a, doneA := newTestRemote(t, &fakeLoL{phase: "Lobby"})
	defer doneA()
	b, doneB := newTestRemote(t, &fakeLoL{phase: "Lobby"})
	defer doneB()
	a.cfg.Port, b.cfg.Port = port, port
	if err := a.startServer(); err != nil {
		t.Fatal(err)
	}
	defer a.stopServer()
	t0 := time.Now()
	if err := b.startServer(); err != nil {
		t.Fatalf("a janela nova devia assumir a porta: %v", err)
	}
	defer b.stopServer()
	a.mu.RLock()
	ret, srv := a.retired, a.srv
	a.mu.RUnlock()
	if !ret || srv != nil {
		t.Fatalf("a janela antiga devia ter soltado (retired=%v srv=%v)", ret, srv != nil)
	}
	t.Logf("assumiu em %v", time.Since(t0))
	// chave errada não derruba ninguém
	c, doneC := newTestRemote(t, &fakeLoL{phase: "Lobby"})
	defer doneC()
	c.cfg.Port, c.cfg.Token = port, "cd"+b.cfg.Token[2:]
	if err := c.startServer(); err == nil {
		t.Fatal("chave errada não devia conseguir tirar a porta")
	}
	if b.srv == nil {
		t.Fatal("a janela certa caiu")
	}
}
