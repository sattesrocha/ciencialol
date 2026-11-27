package main

import (
	"net/http"
	"os"
	"testing"
	"time"
)

// só roda com REMOTE_DEMO=1: sobe o servidor do celular com um LoL simulado por 60 s (pra testar a página)
func TestRemoteDemoServer(t *testing.T) {
	if os.Getenv("REMOTE_DEMO") != "1" {
		t.Skip()
	}
	f := &fakeLoL{phase: "ChampSelect", actType: "pick", timeLeft: 25000}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.PickList = []int{22}
	r.cfg.Ntfy = rmNtfy{On: true, Topic: "clol-abcdefghijklmnopqrstuvwx", Reply: "clolr-x"}
	r.setCiencia(&rmCiencia{Key: 99, Champ: "Lux", Nome: "Lux Tanque", Payload: "{}"})
	go func() {
		for {
			r.tick()
			time.Sleep(500 * time.Millisecond)
		}
	}()
	srv := &http.Server{Addr: "127.0.0.1:8799", Handler: r.handler()}
	go srv.ListenAndServe()
	time.Sleep(25 * time.Second)
}
