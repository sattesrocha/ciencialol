package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// cliente do LoL de mentira, com as mesmas rotas
type fakeLCU struct {
	mu       sync.Mutex
	pages    []perkPage
	maxPages int
	inSelect bool
	spells   [2]int
	sets     map[string]any
	log      []string
}

func (f *fakeLCU) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if u, p, ok := r.BasicAuth(); !ok || u != "riot" || p != "senha" {
			w.WriteHeader(401)
			return
		}
		f.log = append(f.log, r.Method+" "+r.URL.Path)
		js := func(v any) { json.NewEncoder(w).Encode(v) }
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.URL.Path == "/lol-perks/v1/pages" && r.Method == "GET":
			js(f.pages)
		case r.URL.Path == "/lol-perks/v1/pages" && r.Method == "POST":
			n := 0
			for _, p := range f.pages {
				if p.IsDeletable {
					n++
				}
			}
			if n >= f.maxPages {
				w.WriteHeader(400)
				js(map[string]string{"message": "Max pages reached"})
				return
			}
			var in map[string]any
			json.Unmarshal(body, &in)
			if len(in["selectedPerkIds"].([]any)) != 9 {
				w.WriteHeader(400)
				return
			}
			for i := range f.pages {
				f.pages[i].Current = false
			}
			f.pages = append(f.pages, perkPage{ID: 900 + len(f.pages), Name: in["name"].(string), IsDeletable: true, IsEditable: true, Current: true})
			js(f.pages[len(f.pages)-1])
		case strings.HasPrefix(r.URL.Path, "/lol-perks/v1/pages/") && r.Method == "DELETE":
			id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/lol-perks/v1/pages/"))
			var k []perkPage
			for _, p := range f.pages {
				if p.ID != id {
					k = append(k, p)
				}
			}
			f.pages = k
			w.WriteHeader(204)
		case r.URL.Path == "/lol-perks/v1/currentpage":
			for _, p := range f.pages {
				if p.Current {
					js(p)
					return
				}
			}
			w.WriteHeader(404)
		case r.URL.Path == "/lol-champ-select/v1/session":
			if !f.inSelect {
				w.WriteHeader(404)
				return
			}
			js(map[string]any{"localPlayerCellId": 2, "myTeam": []map[string]int{{"cellId": 1, "spell1Id": 4, "spell2Id": 14}, {"cellId": 2, "spell1Id": 12, "spell2Id": 4}}})
		case r.URL.Path == "/lol-champ-select/v1/session/my-selection":
			var in map[string]int
			json.Unmarshal(body, &in)
			f.spells = [2]int{in["spell1Id"], in["spell2Id"]}
			w.WriteHeader(204)
		case r.URL.Path == "/lol-summoner/v1/current-summoner":
			js(map[string]any{"summonerId": 123456789012})
		case r.URL.Path == "/lol-item-sets/v1/item-sets/123456789012/sets":
			if r.Method == "GET" {
				js(f.sets)
				return
			}
			var in map[string]any
			json.Unmarshal(body, &in)
			f.sets = in
			w.WriteHeader(201)
		default:
			w.WriteHeader(404)
		}
	})
}

const testBuild = `{"page":{"primaryStyleId":8100,"subStyleId":8000,"selectedPerkIds":[8112,8143,8138,8106,9111,8014,5008,5008,5011]},
 "spells":[14,4],"champKey":1,"champ":"Annie","title":"Annie Teste",
 "blocks":[{"type":"Inicial","items":["1056","2003"]},{"type":"Principal","items":["3020","6655","x"]}]}`

func TestLcuImport(t *testing.T) {
	f := &fakeLCU{maxPages: 3, inSelect: true,
		pages: []perkPage{{ID: 1, Name: "Minha", IsDeletable: true, Current: true}, {ID: 2, Name: lcuPageName, IsDeletable: true}, {ID: 50, Name: "Padrão", IsDeletable: false}},
		sets:  map[string]any{"accountId": 1, "itemSets": []any{map[string]any{"uid": "outro", "title": "Meu set"}, map[string]any{"uid": "ciencialol-1", "title": "velho"}}}}
	s := httptest.NewTLSServer(f.handler(t))
	defer s.Close()
	c := newLcuConn(s.URL, "senha")
	r, err := lcuImportWith(c, testBuild, false)
	if err != nil || r.Runes != "ok" || r.Spells != "ok" || r.Items != "ok" {
		t.Fatalf("resultado %+v %v", r, err)
	}
	// página antiga do app trocada; a do usuário intacta
	names := []string{}
	for _, p := range f.pages {
		names = append(names, p.Name)
	}
	if strings.Join(names, ",") != "Minha,Padrão,"+lcuPageName {
		t.Fatalf("páginas: %v", names)
	}
	// Flash continua na tecla F (spell2), onde a pessoa já usa
	if f.spells != [2]int{14, 4} {
		t.Fatalf("feitiços: %v", f.spells)
	}
	sets := f.sets["itemSets"].([]any)
	if len(sets) != 2 || sets[1].(map[string]any)["uid"] != "outro" || sets[0].(map[string]any)["title"] != "(Ciência) Annie Teste" {
		t.Fatalf("sets: %v", sets)
	}
	bl := sets[0].(map[string]any)["blocks"].([]any)
	if len(bl) != 2 || len(bl[1].(map[string]any)["items"].([]any)) != 2 {
		t.Fatalf("blocos: %v", bl)
	}
}

func TestLcuPagesFull(t *testing.T) {
	f := &fakeLCU{maxPages: 2, pages: []perkPage{{ID: 1, Name: "A", IsDeletable: true, Current: true}, {ID: 2, Name: "B", IsDeletable: true}}, sets: map[string]any{}}
	s := httptest.NewTLSServer(f.handler(t))
	defer s.Close()
	c := newLcuConn(s.URL, "senha")
	r, _ := lcuImportWith(c, testBuild, false)
	if r.Runes != "full:A" || r.Spells != "fora" {
		t.Fatalf("sem espaço: %+v", r)
	}
	r, _ = lcuImportWith(c, testBuild, true)
	if r.Runes != "ok" || len(f.pages) != 2 || f.pages[1].Name != lcuPageName {
		t.Fatalf("trocar atual: %+v %v", r, f.pages)
	}
}

func TestParseLockfile(t *testing.T) {
	p, s, err := parseLockfile([]byte("LeagueClient:1234:54321:abcDEF:https"))
	if err != nil || p != "54321" || s != "abcDEF" {
		t.Fatal(p, s, err)
	}
	if _, _, err := parseLockfile([]byte("lixo")); err == nil {
		t.Fatal("aceitou lixo")
	}
}

func TestLcuFlashKey(t *testing.T) {
	for _, tc := range []struct {
		key  string
		want [2]int
	}{{"keep", [2]int{14, 4}}, {"", [2]int{14, 4}}, {"d", [2]int{4, 14}}, {"f", [2]int{14, 4}}} {
		f := &fakeLCU{maxPages: 5, inSelect: true, sets: map[string]any{}}
		s := httptest.NewTLSServer(f.handler(t))
		c := newLcuConn(s.URL, "senha")
		b := strings.Replace(testBuild, `"spells":[14,4]`, `"spells":[4,14],"flash":"`+tc.key+`"`, 1)
		if r, _ := lcuImportWith(c, b, false); r.Spells != "ok" || f.spells != tc.want {
			t.Fatalf("flash %q: %v %+v", tc.key, f.spells, r)
		}
		s.Close()
	}
}

func TestLcuRecent(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/lol-summoner/v1/current-summoner":
			w.Write([]byte(`{"puuid":"ME","summonerId":1}`))
		case "/lol-match-history/v1/products/lol/current-summoner/matches":
			w.Write([]byte(`{"games":{"games":[{"gameId":99,"gameCreation":1700000000000,"gameDuration":1800,"queueId":420,"gameMode":"CLASSIC",
			  "participants":[{"participantId":3,"championId":32,"stats":{"win":true,"kills":8,"deaths":3,"assists":12,"totalDamageDealtToChampions":24100,"goldEarned":12300,"totalMinionsKilled":150,"neutralMinionsKilled":30,"visionScore":31,"champLevel":16,"item0":3068}}],
			  "participantIdentities":[{"participantId":3,"player":{"puuid":"ME"}}]}]}}`))
		default:
			w.WriteHeader(404)
		}
	})
	s := httptest.NewTLSServer(h)
	defer s.Close()
	g, err := lcuRecentWith(newLcuConn(s.URL, "x"))
	if err != nil || len(g) != 1 || g[0].Champ != 32 || g[0].K != 8 || g[0].CS != 180 || g[0].Dmg != 24100 || !g[0].Win || g[0].Items[0] != 3068 {
		t.Fatalf("%+v %v", g, err)
	}
}

func TestLcuRecentSlowAndFallback(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/lol-summoner/v1/current-summoner":
			w.Write([]byte(`{"puuid":"ME"}`))
		case "/lol-end-of-game/v1/eog-stats-block":
			w.Write([]byte(`{"gameId":555,"gameLength":1500,"queueId":0,"gameMode":"CLASSIC","localPlayer":{"championId":32,"items":[3068],"stats":{"CHAMPIONS_KILLED":5,"NUM_DEATHS":2,"ASSISTS":9,"TOTAL_DAMAGE_DEALT_TO_CHAMPIONS":18000,"GOLD_EARNED":11000,"MINIONS_KILLED":120,"NEUTRAL_MINIONS_KILLED":10,"VISION_SCORE":20,"LEVEL":15,"WIN":1}}}`))
		default: // histórico indisponível
			w.WriteHeader(404)
		}
	})
	s := httptest.NewTLSServer(h)
	defer s.Close()
	g, err := lcuRecentWith(newLcuConn(s.URL, "x"))
	if err != nil || len(g) != 1 || g[0].GameID != 555 || g[0].K != 5 || g[0].CS != 130 || !g[0].Win {
		t.Fatalf("%+v %v", g, err)
	}
}

func TestLcuBusyThenOk(t *testing.T) {
	// o cliente responde 503 nas primeiras chamadas (fim de partida) e depois volta
	n := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n <= 2 {
			w.WriteHeader(503)
			return
		}
		switch {
		case r.URL.Path == "/lol-perks/v1/pages" && r.Method == "GET":
			w.Write([]byte(`[]`))
		default:
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()
	c := newLcuConn(srv.URL, "x")
	r, err := lcuImportWith(c, `{"page":{"primaryStyleId":8000,"subStyleId":8100,"selectedPerkIds":[1,2,3,4,5,6,7,8,9]}}`, false)
	if err != nil || r.Runes != "ok" {
		t.Fatalf("esperava ok depois de esperar o cliente: %v %+v", err, r)
	}
}

func TestLcuNotResponding(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer srv.Close()
	c := newLcuConn(srv.URL, "x")
	_, err := lcuImportWith(c, `{"page":{"primaryStyleId":8000,"subStyleId":8100,"selectedPerkIds":[1,2,3,4,5,6,7,8,9]}}`, false)
	if err == nil || !strings.Contains(err.Error(), "não está respondendo") {
		t.Fatalf("esperava erro claro, veio %v", err)
	}
}

func TestParseUxArgs(t *testing.T) {
	p, s := parseUxArgs(`"C:/Riot Games/League of Legends/LeagueClientUx.exe" "--riotclient-auth-token=abc" "--app-port=55549" "--remoting-auth-token=Zx-9_k" "--app-pid=1234"`)
	if p != "55549" || s != "Zx-9_k" {
		t.Fatalf("veio %q %q", p, s)
	}
	p, s = parseUxArgs(`LeagueClientUx.exe --app-port=6000 --remoting-auth-token=tok123 --x`)
	if p != "6000" || s != "tok123" {
		t.Fatalf("sem aspas: %q %q", p, s)
	}
}

func TestLcuPhaseConn(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/lol-gameflow/v1/gameflow-phase" {
			w.Write([]byte(`"EndOfGame"`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()
	phaseConn = newLcuConn(srv.URL, "x")
	ph, err := lcuPhase()
	if err != nil || ph != "EndOfGame" {
		t.Fatalf("veio %q %v", ph, err)
	}
}

func TestItemSetFirst(t *testing.T) {
	var put map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/lol-summoner/v1/current-summoner":
			w.Write([]byte(`{"summonerId":7}`))
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/sets"):
			w.Write([]byte(`{"accountId":7,"itemSets":[{"uid":"blitz-1","title":"Blitz: AP Viktor","sortrank":500,"type":"custom"}],"timestamp":1}`))
		case r.Method == "PUT":
			json.NewDecoder(r.Body).Decode(&put)
			w.WriteHeader(201)
		default:
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()
	c := newLcuConn(srv.URL, "x")
	_, err := lcuImportWith(c, `{"champKey":99,"champ":"Lux","title":"Ciência","blocks":[{"type":"Inicial","items":["1056"]}]}`, false)
	if err != nil {
		t.Fatal(err)
	}
	sets := put["itemSets"].([]any)
	first := sets[0].(map[string]any)
	if first["uid"] != "ciencialol-99" || first["sortrank"].(float64) >= 0 || !strings.HasPrefix(first["title"].(string), "(Ciência) ") {
		t.Fatalf("o nosso conjunto não ficou primeiro: %+v", first)
	}
}

func TestLcuImportNeverHangs(t *testing.T) {
	// o cliente aceita a conexão mas nunca responde (como às vezes depois da partida)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(30 * time.Second):
		}
	}))
	defer srv.Close()
	c := newLcuConn(srv.URL, "x")
	c.deadline = time.Now().Add(3 * time.Second)
	c.cl.Timeout = 2 * time.Second
	t0 := time.Now()
	_, err := lcuImportWith(c, `{"page":{"primaryStyleId":8000,"subStyleId":8100,"selectedPerkIds":[1,2,3,4,5,6,7,8,9]},"champKey":1,"blocks":[{"type":"x","items":["1"]}]}`, false)
	if took := time.Since(t0); took > 6*time.Second {
		t.Fatalf("demorou %v (devia respeitar o prazo)", took)
	}
	if err == nil {
		t.Fatalf("esperava erro de tempo")
	}
}

func TestLcuImportParallelSlowServer(t *testing.T) {
	// cada pedido demora 1,5 s (servidor da Riot lento): em fila passaria de 10 s
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		switch {
		case r.URL.Path == "/lol-summoner/v1/current-summoner":
			w.Write([]byte(`{"summonerId":7}`))
		case r.URL.Path == "/lol-perks/v1/pages" && r.Method == "GET":
			w.Write([]byte(`[{"id":5,"name":"Ciência LOL","isDeletable":true}]`))
		case r.URL.Path == "/lol-champ-select/v1/session":
			w.Write([]byte(`{"localPlayerCellId":0,"myTeam":[{"cellId":0,"spell1Id":4,"spell2Id":14}]}`))
		case strings.HasSuffix(r.URL.Path, "/sets") && r.Method == "GET":
			w.Write([]byte(`{"itemSets":[]}`))
		default:
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()
	c := newLcuConn(srv.URL, "x")
	t0 := time.Now()
	r, err := lcuImportWith(c, `{"page":{"primaryStyleId":8000,"subStyleId":8100,"selectedPerkIds":[1,2,3,4,5,6,7,8,9]},"spells":[4,14],"champKey":1,"blocks":[{"type":"x","items":["1"]}]}`, false)
	took := time.Since(t0)
	if err != nil || r.Runes != "ok" || r.Spells != "ok" || r.Items != "ok" {
		t.Fatalf("veio %v %+v", err, r)
	}
	if took > 7*time.Second {
		t.Fatalf("demorou %v: as partes deviam ir ao mesmo tempo", took)
	}
	t.Logf("importou tudo em %v", took)
}
