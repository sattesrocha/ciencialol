package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// imita o Apps Script: o POST responde com redirecionamento 302 pra uma página que devolve o JSON (como o Google faz)
func fakeScript(t *testing.T) *httptest.Server {
	var mu sync.Mutex
	store := map[string]map[string]any{}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path == "/echo" {
			w.Write([]byte(r.URL.Query().Get("o")))
			return
		}
		reply := func(o any) {
			b, _ := json.Marshal(o)
			http.Redirect(w, r, srv.URL+"/echo?o="+urlq(string(b)), http.StatusFound)
		}
		if r.Method == "GET" {
			if r.URL.Query().Get("k") != CloudKey {
				reply(map[string]any{"ok": false, "erro": "senha"})
				return
			}
			reply(map[string]any{"ok": true, "dados": store})
			return
		}
		b, _ := io.ReadAll(r.Body)
		var in map[string]string
		json.Unmarshal(b, &in)
		if in["k"] != CloudKey {
			reply(map[string]any{"ok": false, "erro": "senha"})
			return
		}
		store[in["chave"]] = map[string]any{"ts": 1, "dados": in["dados"]}
		reply(map[string]any{"ok": true})
	}))
	return srv
}

func urlq(s string) string {
	r := strings.NewReplacer("%", "%25", "&", "%26", "+", "%2B", "#", "%23", " ", "%20")
	return r.Replace(s)
}

func TestCloudRoundTrip(t *testing.T) {
	s := fakeScript(t)
	defer s.Close()
	if got, err := cloudGetFrom(s.URL + "/exec"); err != nil || got != "{}" {
		t.Fatalf("vazio: %q %v", got, err)
	}
	if err := cloudPutTo(s.URL+"/exec", "m/i123", `{"a":1}`); err != nil {
		t.Fatal(err)
	}
	got, err := cloudGetFrom(s.URL + "/exec")
	if err != nil || !strings.Contains(got, `"m/i123"`) || !strings.Contains(got, `{\"a\":1}`) {
		t.Fatalf("depois: %q %v", got, err)
	}
	if err := cloudPutTo(s.URL+"/exec", "x/../bad", "{}"); err == nil {
		t.Fatal("chave ruim aceita")
	}
}

func TestCloudAdminKeyRules(t *testing.T) {
	if err := cloudPutAdminTo("http://127.0.0.1:1", "", "c/main", "{}"); err == nil {
		t.Fatal("aceitou sem senha")
	}
	if err := cloudPutAdminTo("http://127.0.0.1:1", "x", "m/abc", "{}"); err == nil {
		t.Fatal("aceitou chave que não é de config")
	}
}
