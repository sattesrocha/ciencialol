package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const fakeGames = `{"data":[
 {"id":"g1","created_at":"2026-10-26T21:00:00+09:00","game_type":"SOLORANKED",
  "participants":[{"summoner":{"game_name":"Outro"},"champion_id":1,"items":[1,2,3],"stats":{"result":"LOSE"}}],
  "myData":{"summoner":{"game_name":"Jogador Teste"},"champion_id":161,"items":[3020,6655,0,4645,3089,0,3340],
   "spells":[4,14],"rune":{"primary_page_id":8200,"primary_rune_id":8229,"secondary_page_id":8300},
   "stats":{"result":"WIN","kill":7,"death":2,"assist":11}}}]}`

const fakeChamp = `{"data":{"summary":{"win_rate":0.51},
 "runes":[{"primary_page_id":8200,"primary_rune_ids":[8229,8226,8210,8237],"secondary_page_id":8300,"secondary_rune_ids":[8345,8347],"stat_mod_ids":[5008,5008,5011],"pick_rate":0.4}],
 "summoner_spells":[{"ids":[4,14],"pick_rate":0.8}],
 "starter_items":[{"ids":[1056,2003]}],"boots":[{"ids":[3020]}],
 "core_items":[{"ids":[6655,4645,3089],"pick_rate":0.3}]}}`

// imita o servidor MCP do op.gg (responde em "event stream", como os servidores MCP fazem)
func fakeMCP(t *testing.T, sse bool) (*httptest.Server, *[]map[string]any) {
	var calls []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(b, &m)
		id := m["id"]
		var result any
		switch m["method"] {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "s1")
			result = map[string]any{"protocolVersion": "2025-03-26"}
		case "notifications/initialized":
			w.WriteHeader(202)
			return
		case "tools/list":
			if r.Header.Get("Mcp-Session-Id") != "s1" {
				w.WriteHeader(400)
				return
			}
			result = json.RawMessage(`{"tools":[
			 {"name":"lol_list_summoner_matches","inputSchema":{"properties":{"game_name":{"type":"string"},"tag_line":{"type":"string"},"region":{"type":"string","enum":["KR","NA","BR","EUW"]},"lang":{"type":"string"},"desired_output_fields":{"type":"array"}},"required":["game_name","tag_line","region"]}},
			 {"name":"lol_get_champion_analysis","inputSchema":{"properties":{"champion":{"type":"string","enum":["AHRI","VELKOZ","MISS_FORTUNE"]},"position":{"type":"string","enum":["TOP","JUNGLE","MID","ADC","SUPPORT"]},"game_mode":{"type":"string","enum":["RANKED","ARAM"]},"lang":{"type":"string"}},"required":["champion","position","game_mode"]}},
			 {"name":"lol_summoner_renewal","inputSchema":{"properties":{"game_name":{"type":"string"},"tag_line":{"type":"string"},"region":{"type":"string"}}}}]}`)
		case "tools/call":
			p := m["params"].(map[string]any)
			calls = append(calls, p)
			txt := `{"ok":true}`
			switch p["name"] {
			case "lol_list_summoner_matches":
				txt = fakeGames
			case "lol_get_champion_analysis":
				txt = fakeChamp
			}
			result = map[string]any{"content": []any{map[string]any{"type": "text", "text": txt}}}
		}
		out, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
		if sse {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Write([]byte("event: message\ndata: " + string(out) + "\n\n"))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(out)
		}
	}))
	return srv, &calls
}

func TestOpggMatchesAndRec(t *testing.T) {
	for _, sse := range []bool{true, false} {
		s, calls := fakeMCP(t, sse)
		old := opggURL
		opggURL = s.URL
		opggTools = nil
		g, err := opggMatches("Jogador Teste#br1")
		if err != nil || len(g) != 1 {
			t.Fatalf("partidas: %v %+v", err, g)
		}
		b := g[0]
		if b.ChampKey != 161 || len(b.Items) != 5 || b.Win == nil || !*b.Win || b.K != 7 || b.Primary != 8200 || len(b.Spells) != 2 || b.At == 0 || b.Mode != "SOLORANKED" {
			t.Fatalf("build: %+v", b)
		}
		var mc map[string]any
		for _, c := range *calls {
			if c["name"] == "lol_list_summoner_matches" {
				mc = c["arguments"].(map[string]any)
			}
		}
		if mc["region"] != "BR" || mc["game_name"] != "Jogador Teste" || mc["tag_line"] != "br1" {
			t.Fatalf("args: %v", mc)
		}
		r, err := opggChampBuild("Vel'Koz", "Velkoz", "sup")
		if err != nil || r.Primary != 8200 || len(r.Perks) != 9 || len(r.Items) != 3 || len(r.Boots) != 1 || len(r.Start) != 2 || len(r.Spells) != 2 {
			t.Fatalf("rec: %v %+v", err, r)
		}
		ca := (*calls)[len(*calls)-1]["arguments"].(map[string]any)
		if ca["champion"] != "VELKOZ" || ca["position"] != "SUPPORT" || ca["game_mode"] != "RANKED" {
			t.Fatalf("args champ: %v", ca)
		}
		opggURL = old
		s.Close()
	}
	if _, err := opggMatches("semtag"); err == nil || !strings.Contains(err.Error(), "Riot ID") {
		t.Fatal("aceitou Riot ID ruim")
	}
}

func TestLcuMyBuilds(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/lol-summoner/v1/current-summoner"):
			w.Write([]byte(`{"puuid":"me"}`))
		case strings.HasPrefix(r.URL.Path, "/lol-match-history/"):
			w.Write([]byte(`{"games":{"games":[{"gameId":9,"gameCreation":1700000000000,"queueId":420,"gameMode":"CLASSIC",
			"participantIdentities":[{"participantId":1,"player":{"puuid":"me"}}],
			"participants":[{"participantId":1,"championId":161,"spell1Id":4,"spell2Id":14,"timeline":{"lane":"BOTTOM","role":"SUPPORT"},
			"stats":{"win":true,"kills":3,"deaths":1,"assists":20,"item0":3020,"item1":0,"item2":6655,"item6":3364,
			"perk0":8229,"perk1":8226,"perk2":8210,"perk3":8237,"perk4":8345,"perk5":8347,"perkPrimaryStyle":8200,"perkSubStyle":8300,
			"statPerk0":5008,"statPerk1":5008,"statPerk2":5011}}]}]}}`))
		default:
			w.WriteHeader(404)
		}
	})
	s := httptest.NewTLSServer(h)
	defer s.Close()
	g, err := lcuMyBuildsWith(newLcuConn(s.URL, "x"))
	if err != nil || len(g) != 1 {
		t.Fatal(err, g)
	}
	b := g[0]
	if b.ChampKey != 161 || len(b.Items) != 3 || len(b.Perks) != 9 || b.Primary != 8200 || b.Sub != 8300 || len(b.Spells) != 2 || !*b.Win || b.A != 20 {
		t.Fatalf("%+v", b)
	}
}

// resposta real do op.gg (Hecarim), no formato compacto deles
func TestOpggRealFormat(t *testing.T) {
	raw, err := os.ReadFile("testdata/opgg_champ.json")
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		Content []struct{ Text string } `json:"content"`
	}
	json.Unmarshal(raw, &r)
	v, err := parseOpggText(r.Content[0].Text)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := opggFindRec(v)
	if !ok || b.Primary != 8000 || b.Sub != 8200 || len(b.Perks) != 9 || b.Perks[0] != 8010 || len(b.Items) != 4 || b.Items[0] != 3070 ||
		len(b.Boots) != 1 || b.Boots[0] != 3009 || len(b.Start) != 2 || len(b.Spells) != 2 || b.Play != 1 || len(b.Skills) != 15 || b.Skills[0] != "Q" || len(b.Alt) != 1 || b.Alt[0] != "jungle" {
		t.Fatalf("%+v", b)
	}
	// e pelo caminho completo (servidor falso devolvendo esse texto)
	if _, err := parseOpggText(`class A: x,y` + "\n\n" + `A(1,"a\"b")`); err != nil {
		t.Fatal(err)
	}
}
