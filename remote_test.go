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
	timeLeft int64 // quanto faltava no instante "stamp" (como o LoL de verdade: não atualiza sozinho)
	stamp    int64 // ms; 0 = agora
	infinite bool
	actType  string
	locked   []int
	hovered  []int
	tphase   string // fase do timer (padrão BAN_PICK)
	ban1     bool   // como o LoL de verdade às vezes faz: lista de "baníveis" com 1 só
}

func (f *fakeLoL) stampNow() int64 {
	if f.stamp == 0 {
		return time.Now().UnixMilli()
	}
	return f.stamp
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
				"timer":             map[string]any{"adjustedTimeLeftInPhase": f.timeLeft, "internalNowInEpochMs": f.stampNow(), "isInfinite": f.infinite, "phase": map[bool]string{true: f.tphase, false: "BAN_PICK"}[f.tphase != ""]},
			})
		case strings.HasSuffix(r.URL.Path, "pickable-champion-ids"):
			w.Write([]byte(`[99,17,22]`)) // Alistar não está liberado
		case strings.HasSuffix(r.URL.Path, "bannable-champion-ids"):
			if f.ban1 {
				w.Write([]byte(`[99]`))
			} else {
				w.Write([]byte(`[99,17,22,12]`))
			}
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
	if w.Code != 200 || !strings.Contains(w.Header().Get("Set-Cookie"), "clr="+tok) || !strings.Contains(w.Header().Get("Set-Cookie"), "HttpOnly") {
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

// o LoL congela o tempo da seleção: "faltavam 30 s há 25 s" = faltam 5 s agora
func TestRemoteAutoUsesRealTimeLeft(t *testing.T) {
	f := &fakeLoL{phase: "ChampSelect", actType: "ban", timeLeft: 30000, stamp: time.Now().Add(-25 * time.Second).UnixMilli()}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.AutoSelect = true
	r.cfg.Auto.BanList = []int{12}
	r.tick()
	if len(f.locked) != 1 || f.locked[0] != 12 {
		t.Fatalf("faltando 5 s devia banir, baniu %v", f.locked)
	}
	// o mesmo "30 s" mas agora mesmo: não age
	f.locked, f.hovered, f.stamp = nil, nil, 0
	r.acted = map[int]time.Time{}
	r.tick()
	if len(f.locked) != 0 {
		t.Fatalf("faltando 30 s não devia agir: %v", f.locked)
	}
}

func TestRemoteAutoInfiniteTimerAndEmptyListExplained(t *testing.T) {
	f := &fakeLoL{phase: "ChampSelect", actType: "ban", timeLeft: 0, infinite: true}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.AutoSelect = true
	r.cfg.Auto.BanList = []int{12}
	r.tick()
	if len(f.locked) != 0 {
		t.Fatal("seleção sem tempo: o automático não devia agir")
	}
	if l := strings.Join(r.logs, "\n"); !strings.Contains(l, "não tem tempo") {
		t.Fatalf("devia explicar no registro: %s", l)
	}
	r2, done2 := newTestRemote(t, &fakeLoL{phase: "ChampSelect", actType: "ban", timeLeft: 20000})
	defer done2()
	r2.cfg.Auto.AutoSelect = true
	r2.tick()
	if l := strings.Join(r2.logs, "\n"); !strings.Contains(l, "lista de bans está vazia") {
		t.Fatalf("devia avisar da lista vazia: %s", l)
	}
}

func TestTimeLeftNow(t *testing.T) {
	now := time.UnixMilli(1_000_000)
	if v := timeLeftNow(30000, 1_000_000-25000, now); v != 5 {
		t.Fatalf("esperava 5, veio %v", v)
	}
	if v := timeLeftNow(30000, 0, now); v != 30 {
		t.Fatalf("sem instante devia usar 30, veio %v", v)
	}
	if v := timeLeftNow(3000, 1_000_000-9000, now); v != 0 {
		t.Fatalf("não pode ficar negativo, veio %v", v)
	}
}

// visto num jogo de verdade: na vez de banir o LoL devolveu só 1 "banível"; e no planejamento a ação
// de ban já aparece "em andamento", mas não dá pra banir ainda
func TestRemoteAutoBanListaDoLoLComUmSo(t *testing.T) {
	f := &fakeLoL{phase: "ChampSelect", actType: "ban", timeLeft: 3000, ban1: true, tphase: "PLANNING"}
	r, done := newTestRemote(t, f)
	defer done()
	r.cfg.Auto.AutoSelect, r.cfg.Auto.BanList, r.cfg.Auto.UseCiencia = true, []int{22}, false
	r.tick()
	r.tick() // carrega a lista de campeões
	if len(f.locked) != 0 {
		t.Fatalf("no planejamento não pode banir: %v", f.locked)
	}
	f.mu.Lock()
	f.tphase = "BAN_PICK"
	f.mu.Unlock()
	r.tick()
	if len(f.locked) != 1 || f.locked[0] != 22 {
		t.Fatalf("devia banir a Ashe (22) mesmo com o LoL dizendo que só dá pra banir 1: %v", f.locked)
	}
}

// pedido de mudança que não traz as listas (ex.: só os segundos) não pode apagar bans/picks
func TestSetAutoSemListasNaoApaga(t *testing.T) {
	r, done := newTestRemote(t, &fakeLoL{phase: "Lobby"})
	defer done()
	r.setAuto(rmAuto{AutoSelect: true, SecondsLeft: 8, BanList: []int{53}, PickList: []int{145}})
	var a rmAuto
	json.Unmarshal([]byte(`{"autoSelect":true,"secondsLeft":18}`), &a)
	got := r.setAuto(a)
	if len(got.BanList) != 1 || len(got.PickList) != 1 {
		t.Fatalf("apagou as listas: %+v", got)
	}
	json.Unmarshal([]byte(`{"autoSelect":true,"secondsLeft":18,"banList":[],"pickList":[145]}`), &a)
	if got := r.setAuto(a); len(got.BanList) != 0 {
		t.Fatalf("lista vazia de propósito devia apagar: %+v", got)
	}
}
