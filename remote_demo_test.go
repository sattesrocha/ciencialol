package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

// só roda com REMOTE_DEMO=1: sobe o servidor do celular com um LoL simulado (pra testar a página)
// REMOTE_DEMO_PHASE escolhe a fase (ChampSelect, ReadyCheck, Matchmaking...)
func TestRemoteDemoServer(t *testing.T) {
	if os.Getenv("REMOTE_DEMO") != "1" {
		t.Skip()
	}
	ph := os.Getenv("REMOTE_DEMO_PHASE")
	if ph == "" {
		ph = "ChampSelect"
	}
	f := &fakeLoL{phase: ph, actType: "pick", timeLeft: 25000}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.PickList = []int{22}
	r.cfg.Auto.BanList = []int{12, 17}
	r.cfg.Enabled = true
	r.cfg.Ntfy = rmNtfy{On: true, Topic: "clol-abcdefghijklmnopqrstuvwx", Reply: "clolr-x"}
	r.setDDChamps("15.24.1", []rmChamp{{ID: 99, Name: "Lux", Alias: "Lux"}, {ID: 17, Name: "Teemo", Alias: "Teemo"}, {ID: 22, Name: "Ashe", Alias: "Ashe"}, {ID: 12, Name: "Alistar", Alias: "Alistar"}})
	r.setCiencia(&rmCiencia{Key: 99, Champ: "Lux", Nome: "Lux Tanque", Payload: "{}"})
	go func() {
		for {
			r.tick()
			time.Sleep(500 * time.Millisecond)
		}
	}()
	if ns := os.Getenv("REMOTE_DEMO_NTFY"); ns != "" { // ponte pelo ntfy (ntfy de mentira)
		r.cfg.Ntfy.Server = ns
		go r.relayLoop(context.Background())
	}
	srv := &http.Server{Addr: "127.0.0.1:" + os.Getenv("REMOTE_DEMO_PORT"), Handler: r.handler()}
	go srv.ListenAndServe()
	time.Sleep(90 * time.Second)
}
