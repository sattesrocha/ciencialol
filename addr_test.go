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

func TestStableLinkAndAddrPublish(t *testing.T) {
	var mu sync.Mutex
	got := map[string]string{}
	ns := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		got[r.URL.Path] = string(b)
		mu.Unlock()
	}))
	defer ns.Close()
	r := &remoteApp{}
	r.cfg = rmConfig{Port: 8787, Token: strings.Repeat("ab", 16), AddrTopic: "clola-abcdefghij", Ntfy: rmNtfy{Server: ns.URL}}
	l := r.stableLink("i")
	if !strings.HasPrefix(l, stableBase+"#clola-abcdefghij."+strings.Repeat("ab", 16)) || !strings.HasSuffix(l, "&i") {
		t.Fatalf("link fixo errado: %s", l)
	}
	if !strings.Contains(l, "&s=") { // servidor próprio vai junto
		t.Fatalf("faltou o servidor: %s", l)
	}
	if err := r.addrPublish([]string{"192.168.0.10"}, 8787); err != nil {
		t.Fatal(err)
	}
	var m addrMsg
	json.Unmarshal([]byte(got["/clola-abcdefghij"]), &m)
	if len(m.IPs) != 1 || m.IPs[0] != "192.168.0.10" || m.Port != 8787 {
		t.Fatalf("caixa postal errada: %+v", got)
	}
	if strings.Contains(got["/clola-abcdefghij"], strings.Repeat("ab", 16)) {
		t.Fatal("a chave de acesso NÃO pode ir pra caixa postal")
	}
}
