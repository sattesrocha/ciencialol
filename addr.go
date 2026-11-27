package main

// Endereço fixo pro celular: o atalho no celular aponta pro site do Ciência LOL (que nunca muda),
// e o site descobre onde está o PC agora lendo uma "caixa postal" no ntfy. O PC anota ali o seu
// IP da rede de casa ao ligar, quando o IP muda e a cada 3 h. Na caixa vai SÓ o IP de casa
// (inútil fora da casa); a chave de acesso fica no próprio atalho do celular, numa parte do link
// (depois do #) que o navegador nunca manda pra nenhum servidor.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const stableBase = "https://sattesrocha.github.io/ciencialol/celular/"

type addrMsg struct {
	IPs  []string `json:"ips"`
	Port int      `json:"port"`
	TS   int64    `json:"ts"`
}

func (r *remoteApp) addrEnsureTopic() {
	r.mu.Lock()
	if r.cfg.AddrTopic == "" {
		r.cfg.AddrTopic = randTopic("clola-")
	}
	r.mu.Unlock()
}

// link fixo: #caixa.chave (+ "&a" = dica de criar atalho, "&n" = ir pra parte do ntfy)
func (r *remoteApp) stableLink(mode string) string {
	r.mu.RLock()
	topic, tok, srv := r.cfg.AddrTopic, r.cfg.Token, r.cfg.Ntfy.Server
	r.mu.RUnlock()
	if topic == "" {
		return ""
	}
	l := stableBase + "#" + topic + "." + tok
	if srv != "" && srv != ntfyDefaultServer {
		l += "&s=" + url.QueryEscape(srv)
	}
	if mode != "" {
		l += "&" + mode
	}
	return l
}

func (r *remoteApp) addrPublish(ips []string, port int) error {
	n := r.ntfyCfg()
	r.mu.RLock()
	topic := r.cfg.AddrTopic
	r.mu.RUnlock()
	if topic == "" || len(ips) == 0 {
		return nil
	}
	body, _ := json.Marshal(addrMsg{IPs: ips, Port: port, TS: time.Now().Unix()})
	req, _ := http.NewRequest("POST", n.Server+"/"+topic, bytes.NewReader(body))
	req.Header.Set("Cache", "yes")
	req.Header.Set("Firebase", "no") // não é aviso: não repassa pro Google/Apple
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}
	resp, err := ntfyClient.Do(req)
	if err != nil {
		return fmt.Errorf("sem conexão com o ntfy (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("o ntfy recusou (%d): %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

// enquanto o controle pelo celular estiver ligado: mantém a caixa postal em dia
func (r *remoteApp) addrLoop(ctx context.Context) {
	last, lastAt, failed := "", time.Time{}, false
	for {
		r.mu.RLock()
		port := r.cfg.Port
		r.mu.RUnlock()
		ips := localIPv4s()
		key := strings.Join(ips, ",")
		if len(ips) > 0 && (key != last || time.Since(lastAt) > 3*time.Hour) {
			if err := r.addrPublish(ips, port); err == nil {
				if last != "" && key != last {
					r.addLog("O IP do PC mudou: o atalho do celular já sabe o novo")
				} else if last == "" {
					r.addLog("Endereço do PC anotado pro atalho do celular ✓ (" + ips[0] + ")")
				}
				last, lastAt, failed = key, time.Now(), false
			} else if !failed {
				r.addLog("Não deu pra anotar o endereço do PC: " + err.Error())
				failed = true
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
	}
}
