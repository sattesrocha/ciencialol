package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLcuGameWith(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lol-match-history/v1/games/123" {
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`{"gameId":123,"gameDuration":1800,"gameMode":"CLASSIC","queueId":420,
		"participants":[
		 {"participantId":1,"teamId":100,"championId":99,"spell1Id":4,"spell2Id":14,"stats":{"win":true,"kills":7,"deaths":2,"assists":9,"totalMinionsKilled":150,"neutralMinionsKilled":10,"totalDamageDealtToChampions":25000,"goldEarned":12000,"visionScore":20,"champLevel":16,"item0":3089,"item6":3340,"perk0":8112,"perkSubStyle":8300}},
		 {"participantId":2,"teamId":200,"championId":12,"spell1Id":4,"spell2Id":3,"stats":{"win":false,"kills":1,"deaths":8,"assists":3}}],
		"participantIdentities":[{"participantId":1,"player":{"gameName":"Fulano","tagLine":"BR1"}},{"participantId":2,"player":{"summonerName":"Ciclano"}}]}`))
	}))
	defer srv.Close()
	g, err := lcuGameWith(newLcuConn(srv.URL, "x"), 123)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Teams) != 2 || !g.Teams[0].Win || g.Teams[1].Win {
		t.Fatalf("times errados: %+v", g.Teams)
	}
	p := g.Teams[0].Players[0]
	if p.Name != "Fulano" || p.Champ != 99 || p.CS != 160 || p.Spells != [2]int{4, 14} || p.Items[0] != 3089 || p.Rune != 8112 || p.Sub != 8300 {
		t.Fatalf("jogador errado: %+v", p)
	}
	if g.Teams[1].Players[0].Name != "Ciclano" {
		t.Fatalf("nome antigo não usado: %+v", g.Teams[1].Players[0])
	}
	if _, err := lcuGameWith(newLcuConn(srv.URL, "x"), 999); err == nil {
		t.Fatal("esperava erro de partida não encontrada")
	}
}
