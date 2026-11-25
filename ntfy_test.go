package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// ntfy simulado: guarda o que foi publicado e entrega respostas no tópico de respostas
type fakeNtfy struct {
	mu      sync.Mutex
	pub     []ntfyMsg
	replies chan string
}

func (n *fakeNtfy) srv() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/" {
			var m ntfyMsg
			json.NewDecoder(r.Body).Decode(&m)
			n.mu.Lock()
			n.pub = append(n.pub, m)
			n.mu.Unlock()
			return
		}
		if strings.HasSuffix(r.URL.Path, "/json") {
			w.(http.Flusher).Flush()
			for {
				select {
				case b := <-n.replies:
					fmt.Fprintf(w, `{"id":"x%d","event":"message","message":%q}`+"\n", time.Now().UnixNano(), b)
					w.(http.Flusher).Flush()
				case <-r.Context().Done():
					return
				}
			}
		}
	}))
}

func (n *fakeNtfy) last() ntfyMsg {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.pub) == 0 {
		return ntfyMsg{}
	}
	return n.pub[len(n.pub)-1]
}

func waitFor(t *testing.T, cond func() bool, what string) {
	for i := 0; i < 100; i++ {
		if cond() {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatal("esperando: " + what)
}

func TestNtfyAcceptAndPickFromNotification(t *testing.T) {
	f := &fakeLoL{phase: "ReadyCheck"}
	r, done := newTestRemote(t, f)
	defer done()
	nf := &fakeNtfy{replies: make(chan string, 4)}
	ns := nf.srv()
	defer ns.Close()
	r.cfg.Ntfy = rmNtfy{On: true, Server: ns.URL, Topic: "clol-teste", Reply: "clolr-teste"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.ntfyListen(ctx)

	// fila estourou: notificação com Aceitar/Recusar, só uma vez por fila
	r.tick()
	r.tick()
	waitFor(t, func() bool { return nf.last().Title == "Partida encontrada!" }, "aviso da fila")
	m := nf.last()
	if len(m.Actions) < 2 || m.Actions[0].URL != ns.URL+"/clolr-teste" || !strings.HasPrefix(m.Actions[0].Body, "a:") {
		t.Fatalf("botões errados: %+v", m.Actions)
	}
	nf.mu.Lock()
	if n := len(nf.pub); n != 1 {
		t.Fatalf("devia avisar uma vez só, avisou %d", n)
	}
	nf.mu.Unlock()
	// código errado não faz nada; o certo aceita; repetir o certo não faz nada
	nf.replies <- "a:codigo-falso"
	nf.replies <- m.Actions[0].Body
	waitFor(t, func() bool { f.mu.Lock(); defer f.mu.Unlock(); return f.accepted }, "aceitar pela notificação")
	if r.rcCode != "" {
		t.Fatal("o código devia ser de uso único")
	}

	// vez de escolher com a ciência: botão "Escolher Lux"
	f.mu.Lock()
	f.phase, f.actType, f.timeLeft = "ChampSelect", "pick", 25000
	f.mu.Unlock()
	r.setCiencia(&rmCiencia{Key: 99, Champ: "Lux", Nome: "Lux Tanque"})
	r.tick()
	waitFor(t, func() bool { return nf.last().Title == "Sua vez de escolher" }, "aviso da vez")
	m = nf.last()
	if len(m.Actions) == 0 || m.Actions[0].Label != "Escolher Lux" {
		t.Fatalf("faltou o botão da ciência: %+v", m.Actions)
	}
	nf.replies <- m.Actions[0].Body
	waitFor(t, func() bool { f.mu.Lock(); defer f.mu.Unlock(); return len(f.locked) == 1 && f.locked[0] == 99 }, "escolher pela notificação")
}

func TestNtfyOldCodeIsUseless(t *testing.T) {
	f := &fakeLoL{phase: "ReadyCheck"}
	r, done := newTestRemote(t, f)
	defer done()
	nf := &fakeNtfy{replies: make(chan string, 1)}
	ns := nf.srv()
	defer ns.Close()
	r.cfg.Ntfy = rmNtfy{On: true, Server: ns.URL, Topic: "t", Reply: "r"}
	r.tick()
	waitFor(t, func() bool { return nf.last().Title != "" }, "aviso")
	body := nf.last().Actions[0].Body
	f.phase = "Lobby" // a fila acabou
	r.tick()
	r.handleReply(body)
	if f.accepted {
		t.Fatal("código de uma fila que já acabou não pode aceitar")
	}
}

func TestNtfyServerValidation(t *testing.T) {
	for in, want := range map[string]string{"": ntfyDefaultServer, "https://ntfy.sh/": "https://ntfy.sh", "https://meu.servidor": "https://meu.servidor"} {
		if got, err := validServer(in); err != nil || got != want {
			t.Errorf("%q: %q %v", in, got, err)
		}
	}
	for _, bad := range []string{"ftp://x", "ntfy.sh", "https://x/caminho", "javascript:alert(1)"} {
		if _, err := validServer(bad); err == nil {
			t.Errorf("devia recusar %q", bad)
		}
	}
}
