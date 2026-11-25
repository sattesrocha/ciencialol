package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// LoL simulado: fila, seleção com a vez de escolher, e registra o que o app mandou
type fakeLoL struct {
	mu       sync.Mutex
	phase    string
	accepted bool
	timeLeft int64
	actType  string
	locked   []int
	hovered  []int
}

func (f *fakeLoL) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case r.URL.Path == "/lol-gameflow/v1/gameflow-phase":
			json.NewEncoder(w).Encode(f.phase)
		case r.URL.Path == "/lol-game-data/assets/v1/champion-summary.json":
			w.Write([]byte(`[{"id":-1,"name":"None"},{"id":99,"name":"Lux"},{"id":17,"name":"Teemo"},{"id":22,"name":"Ashe"},{"id":12,"name":"Alistar"}]`))
		case r.URL.Path == "/lol-matchmaking/v1/ready-check":
			w.Write([]byte(`{"state":"InProgress","playerResponse":"None","timer":3}`))
		case r.URL.Path == "/lol-matchmaking/v1/ready-check/accept":
			f.accepted = true
		case r.URL.Path == "/lol-champ-select/v1/session":
			json.NewEncoder(w).Encode(map[string]any{
				"localPlayerCellId": 0,
				"myTeam":            []map[string]any{{"cellId": 0}, {"cellId": 1, "championId": 17}},
				"theirTeam":         []map[string]any{{"cellId": 5}},
				"actions":           [][]map[string]any{{{"id": 7, "actorCellId": 0, "isInProgress": true, "type": f.actType}}},
				"timer":             map[string]any{"adjustedTimeLeftInPhase": f.timeLeft, "phase": "BAN_PICK"},
			})
		case strings.HasSuffix(r.URL.Path, "pickable-champion-ids"):
			w.Write([]byte(`[99,17,22]`)) // Alistar não está liberado
		case strings.HasSuffix(r.URL.Path, "bannable-champion-ids"):
			w.Write([]byte(`[99,17,22,12]`))
		case r.Method == "PATCH" && strings.HasPrefix(r.URL.Path, "/lol-champ-select/v1/session/actions/"):
			var b struct{ ChampionID int }
			json.NewDecoder(r.Body).Decode(&b)
			f.hovered = append(f.hovered, b.ChampionID)
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/complete"):
			f.locked = append(f.locked, f.hovered[len(f.hovered)-1])
		default:
			w.WriteHeader(404)
		}
	}
}

func newTestRemote(t *testing.T, f *fakeLoL) (*remoteApp, func()) {
	srv := httptest.NewTLSServer(f.handler(t))
	r := &remoteApp{acted: map[int]time.Time{}, fails: map[string][]time.Time{}}
	r.cfg = rmConfig{Port: 8787, Token: strings.Repeat("ab", 16), Auto: rmAuto{SecondsLeft: 8, UseCiencia: true}}
	r.conn = newLcuConn(srv.URL, "x")
	return r, srv.Close
}

func TestRemoteAutoAcceptAndPickCiencia(t *testing.T) {
	f := &fakeLoL{phase: "ReadyCheck"}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.AutoAccept = true
	r.tick()
	if !f.accepted {
		t.Fatal("devia aceitar a partida sozinho")
	}
	// seleção: falta pouco tempo, ciência sorteada = Lux (99), lista: Alistar (bloqueado), Ashe
	f.phase, f.actType, f.timeLeft = "ChampSelect", "pick", 5000
	r.cfg.Auto.AutoSelect = true
	r.cfg.Auto.PickList = []int{12, 22}
	r.setCiencia(&rmCiencia{Key: 99, Champ: "Lux", Nome: "Lux Tanque"})
	r.tick()
	if len(f.locked) != 1 || f.locked[0] != 99 {
		t.Fatalf("devia travar a Lux da ciência, travou %v", f.locked)
	}
	// sem ciência: pula o Alistar (não liberado) e escolhe a Ashe
	f.locked, f.hovered = nil, nil
	r.setCiencia(nil)
	r.acted = map[int]time.Time{}
	r.tick()
	if len(f.locked) != 1 || f.locked[0] != 22 {
		t.Fatalf("devia travar a Ashe, travou %v", f.locked)
	}
	// ainda com muito tempo: não age
	f.locked, f.hovered, f.timeLeft = nil, nil, 25000
	r.acted = map[int]time.Time{}
	r.tick()
	if len(f.locked) != 0 {
		t.Fatalf("agiu antes da hora: %v", f.locked)
	}
}

func TestRemoteBanNeverBansOwnCiencia(t *testing.T) {
	f := &fakeLoL{phase: "ChampSelect", actType: "ban", timeLeft: 3000}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.AutoSelect = true
	r.cfg.Auto.BanList = []int{99, 12}
	r.setCiencia(&rmCiencia{Key: 99, Champ: "Lux", Nome: "Lux Tanque"})
	r.tick()
	if len(f.locked) != 1 || f.locked[0] != 12 {
		t.Fatalf("devia pular a Lux (ciência) e banir o Alistar, baniu %v", f.locked)
	}
}

func TestRemoteHTTPSecurity(t *testing.T) {
	f := &fakeLoL{phase: "Lobby"}
	r, done := newTestRemote(t, f)
	defer done()
	h := r.handler()
	tok := r.cfg.Token
	call := func(method, target, from string, hdr map[string]string, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.RemoteAddr = from
		for k, v := range hdr {
			req.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}
	// fora da rede local: recusado mesmo com token
	if w := call("GET", "/api/state", "8.8.8.8:5555", map[string]string{"X-Token": tok}, ""); w.Code != 403 {
		t.Fatalf("IP público devia ser recusado, veio %d", w.Code)
	}
	// rede local sem token: 401
	if w := call("GET", "/api/state", "192.168.0.20:5555", nil, ""); w.Code != 401 {
		t.Fatalf("sem token devia dar 401, veio %d", w.Code)
	}
	// Tailscale com token: ok
	if w := call("GET", "/api/state", "100.101.102.103:5555", map[string]string{"X-Token": tok}, ""); w.Code != 200 {
		t.Fatalf("Tailscale com token devia passar, veio %d", w.Code)
	}
	// link com token vira cookie
	w := call("GET", "/?t="+tok, "192.168.0.20:5555", nil, "")
	if w.Code != 302 || !strings.Contains(w.Header().Get("Set-Cookie"), "clr="+tok) || !strings.Contains(w.Header().Get("Set-Cookie"), "HttpOnly") {
		t.Fatalf("link devia gravar cookie: %d %q", w.Code, w.Header().Get("Set-Cookie"))
	}
	// pedido de outro site (Origin diferente) é recusado
	if w := call("POST", "/api/accept", "192.168.0.20:5555", map[string]string{"X-Token": tok, "Origin": "http://evil.example"}, "{}"); w.Code != 403 {
		t.Fatalf("Origin de fora devia ser recusado, veio %d", w.Code)
	}
	// martelar token errado bloqueia
	for i := 0; i < 12; i++ {
		call("GET", "/?t=errado", "192.168.0.66:5555", nil, "")
	}
	if w := call("GET", "/api/state", "192.168.0.66:5555", map[string]string{"X-Token": tok}, ""); w.Code != 429 {
		t.Fatalf("depois de muitas tentativas devia bloquear, veio %d", w.Code)
	}
	// página do celular sai com o cookie certo
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.0.20:5555"
	req.AddCookie(&http.Cookie{Name: "clr", Value: tok})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	b, _ := io.ReadAll(rec.Body)
	if rec.Code != 200 || !strings.Contains(string(b), "Ciência LOL") {
		t.Fatalf("página do celular não abriu: %d", rec.Code)
	}
}

func TestAllowedRemoteIP(t *testing.T) {
	for ip, want := range map[string]bool{"127.0.0.1:1": true, "10.0.0.5:1": true, "172.20.1.1:1": true, "192.168.1.9:1": true,
		"100.64.0.1:1": true, "100.127.255.1:1": true, "100.128.0.1:1": false, "8.8.8.8:1": false, "[fe80::1]:1": true, "[2001:db8::1]:1": false} {
		if got := allowedRemoteIP(ip); got != want {
			t.Errorf("%s: veio %v, esperava %v", ip, got, want)
		}
	}
}
