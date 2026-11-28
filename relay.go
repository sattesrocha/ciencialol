package main

// Ponte pelo ntfy: a tela do celular (no site do Ciência LOL) conversa com o PC pelo ntfy, então
// funciona no Wi-Fi de casa e no 4G. Tudo vai criptografado (AES-GCM) com uma chave tirada do
// próprio atalho do celular: o ntfy só vê um monte de letras. Os nomes dos tópicos também saem
// da chave, então só quem tem o atalho sabe onde procurar.
//
// PC → celular: o estado da partida (só quando muda, e só enquanto tem celular olhando).
// Celular → PC: comandos (aceitar, escolher, travar…), cada um com hora e id (não dá pra repetir).

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type relayKeys struct {
	Down, Up string // tópicos: PC → celular, celular → PC
	key      []byte
}

func relayDerive(token string) relayKeys {
	h := func(label string) []byte {
		m := hmac.New(sha256.New, []byte(token))
		m.Write([]byte(label))
		return m.Sum(nil)
	}
	return relayKeys{
		Down: "clold-" + hex.EncodeToString(h("down"))[:24],
		Up:   "clolu-" + hex.EncodeToString(h("up"))[:24],
		key:  h("enc"),
	}
}

func (k relayKeys) seal(plain []byte, dir string) string {
	blk, _ := aes.NewCipher(k.key)
	g, _ := cipher.NewGCM(blk)
	iv := make([]byte, g.NonceSize())
	rand.Read(iv)
	return base64.StdEncoding.EncodeToString(append(iv, g.Seal(nil, iv, plain, []byte(dir))...))
}

func (k relayKeys) open(msg, dir string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, err
	}
	blk, _ := aes.NewCipher(k.key)
	g, _ := cipher.NewGCM(blk)
	if len(raw) < g.NonceSize()+g.Overhead() {
		return nil, errors.New("curto")
	}
	return g.Open(nil, raw[:g.NonceSize()], raw[g.NonceSize():], []byte(dir))
}

func gz(b []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
}

// o que vai pro celular
type relayOut struct {
	T     string  `json:"t"` // "state" ou "res"
	TS    int64   `json:"ts"`
	State *rmView `json:"state,omitempty"`
	// contagens em hora de relógio (o celular conta sozinho, sem precisar de mensagem por segundo)
	Deadline int64  `json:"deadline,omitempty"` // fim do timer da seleção (ms)
	RCStart  int64  `json:"rcStart,omitempty"`  // quando a partida foi encontrada (ms)
	ID       string `json:"id,omitempty"`
	OK       bool   `json:"ok,omitempty"`
	Err      string `json:"err,omitempty"`
	Data     any    `json:"data,omitempty"`
}

// o que vem do celular
type relayIn struct {
	ID   string          `json:"id"`
	TS   int64           `json:"ts"`
	Cmd  string          `json:"cmd"`
	Body json.RawMessage `json:"body"`
}

type relayRun struct {
	r       *remoteApp
	k       relayKeys
	mu      sync.Mutex
	watch   time.Time // tem celular olhando até…
	force   bool
	seen    map[string]time.Time
	last    []byte
	lastDL  int64
	lastPub time.Time
	failLog bool
	pausa   time.Time // o ntfy pediu calma (429): espera antes de tentar de novo
	lastKey string    // fase + vez + partida encontrada (mudou = manda logo)
}

func (r *remoteApp) relayLoop(ctx context.Context) {
	rr := &relayRun{r: r, k: relayDerive(r.token()), seen: map[string]time.Time{}}
	go rr.listen(ctx)
	for ctx.Err() == nil {
		rr.pump()
		select {
		case <-ctx.Done():
			return
		case <-time.After(400 * time.Millisecond):
		}
	}
}

func (rr *relayRun) publish(o relayOut) error {
	o.TS = time.Now().UnixMilli()
	b, _ := json.Marshal(o)
	n := rr.r.ntfyCfg()
	req, _ := http.NewRequest("POST", n.Server+"/"+rr.k.Down, bytes.NewReader([]byte(rr.k.seal(gz(b), "pc"))))
	req.Header.Set("Firebase", "no")
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}
	resp, err := ntfyClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy %d", resp.StatusCode)
	}
	return nil
}

// manda o estado quando muda (no máximo 1 por segundo) e um sinal de vida a cada 45 s
func (rr *relayRun) pump() {
	rr.mu.Lock()
	watching, force := time.Now().Before(rr.watch), rr.force
	rr.mu.Unlock()
	if !watching {
		return
	}
	v := rr.r.viewNow()
	if len(v.Log) > 12 {
		v.Log = v.Log[len(v.Log)-12:]
	}
	o := relayOut{T: "state"}
	now := time.Now().UnixMilli()
	if v.ChampSelect != nil {
		o.Deadline = now + int64(v.ChampSelect.TimeLeft*1000)
		v.ChampSelect.TimeLeft = 0
	}
	if v.ReadyCheck != nil {
		o.RCStart = now - int64(v.ReadyCheck.Timer*1000)
		v.ReadyCheck.Timer = 0
	}
	cmp, _ := json.Marshal(v)
	o.State = &v
	// O ntfy.sh grátis aguenta ~1 mensagem a cada 5 s (com uma reserva no começo). Mudança importante
	// (fase, sua vez, partida encontrada) vai na hora; o resto no máximo a cada 6 s. O timer o celular
	// conta sozinho. Assim sobra espaço pros avisos de verdade.
	key := v.Phase
	if v.ChampSelect != nil {
		key += "|" + v.ChampSelect.Phase
		if v.ChampSelect.MyAction != nil {
			key += fmt.Sprintf("|%d", v.ChampSelect.MyAction.ID)
		}
	}
	if v.ReadyCheck != nil {
		key += "|rc|" + v.ReadyCheck.PlayerResponse
	}
	rr.mu.Lock()
	if time.Now().Before(rr.pausa) {
		rr.mu.Unlock()
		return
	}
	since := time.Since(rr.lastPub)
	important := key != rr.lastKey
	changed := !bytes.Equal(cmp, rr.last) || abs64(o.Deadline-rr.lastDL) > 2500
	due := (force && since >= 1500*time.Millisecond) || (important && since >= 1500*time.Millisecond) ||
		(changed && since >= 6*time.Second) || since >= 50*time.Second
	if !due {
		rr.mu.Unlock()
		return
	}
	rr.force, rr.lastPub = false, time.Now()
	rr.mu.Unlock()
	if err := rr.publish(o); err != nil {
		rr.mu.Lock()
		rr.force = true // tenta de novo daqui a pouco (sem ficar martelando o ntfy)
		if strings.Contains(err.Error(), "429") {
			rr.pausa = time.Now().Add(40 * time.Second)
		} else {
			rr.pausa = time.Now().Add(5 * time.Second)
		}
		first := !rr.failLog
		rr.failLog = true
		rr.mu.Unlock()
		if first {
			rr.r.addLog("Celular pelo 4G: não deu pra mandar pelo ntfy (" + err.Error() + ")")
		}
		return
	}
	rr.mu.Lock()
	rr.last, rr.lastDL, rr.lastKey, rr.failLog = cmp, o.Deadline, key, false
	rr.mu.Unlock()
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func (rr *relayRun) listen(ctx context.Context) {
	since := fmt.Sprint(time.Now().Unix())
	stream := &http.Client{}
	backoff := 2 * time.Second
	for ctx.Err() == nil {
		n := rr.r.ntfyCfg()
		req, _ := http.NewRequestWithContext(ctx, "GET", n.Server+"/"+rr.k.Up+"/json?since="+since, nil)
		if n.Token != "" {
			req.Header.Set("Authorization", "Bearer "+n.Token)
		}
		resp, err := stream.Do(req)
		if err == nil && resp.StatusCode == 200 {
			backoff = 2 * time.Second
			sc := bufio.NewScanner(resp.Body)
			sc.Buffer(make([]byte, 64*1024), 64*1024)
			for sc.Scan() {
				var m struct {
					ID      string `json:"id"`
					Event   string `json:"event"`
					Message string `json:"message"`
				}
				if json.Unmarshal(sc.Bytes(), &m) != nil {
					continue
				}
				if m.ID != "" {
					since = m.ID
				}
				if m.Event == "message" {
					rr.handle(m.Message)
				}
			}
			resp.Body.Close()
		} else if resp != nil {
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (rr *relayRun) handle(msg string) {
	plain, err := rr.k.open(msg, "cel")
	if err != nil {
		return // não é nosso (ou foi mexido): ignora
	}
	var in relayIn
	if json.Unmarshal(plain, &in) != nil || in.ID == "" {
		return
	}
	now := time.Now()
	if d := now.UnixMilli() - in.TS; d > 120000 || d < -120000 {
		return // velho demais (ou relógio muito errado): não repete comando antigo
	}
	rr.mu.Lock()
	if _, dup := rr.seen[in.ID]; dup {
		rr.mu.Unlock()
		return
	}
	for id, t := range rr.seen {
		if now.Sub(t) > 5*time.Minute {
			delete(rr.seen, id)
		}
	}
	rr.seen[in.ID] = now
	rr.watch = now.Add(10 * time.Minute)
	rr.force = true
	rr.mu.Unlock()
	if in.Cmd == "hello" {
		return
	}
	data, err := rr.r.relayCmd(in.Cmd, in.Body)
	res := relayOut{T: "res", ID: in.ID, OK: err == nil, Data: data}
	if err != nil {
		res.Err = err.Error()
	}
	rr.publish(res)
}

func (r *remoteApp) relayCmd(cmd string, body json.RawMessage) (any, error) {
	var p struct{ ActionID, ChampionID int }
	switch cmd {
	case "accept":
		if _, _, err := r.lcuDo("POST", "/lol-matchmaking/v1/ready-check/accept", nil); err != nil {
			return nil, err
		}
		r.addLog("Celular: partida aceita")
	case "decline":
		if _, _, err := r.lcuDo("POST", "/lol-matchmaking/v1/ready-check/decline", nil); err != nil {
			return nil, err
		}
		r.addLog("Celular: partida recusada")
	case "hover":
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, err
		}
		return nil, r.hover(p.ActionID, p.ChampionID)
	case "lock":
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, err
		}
		if err := r.lock(p.ActionID, p.ChampionID); err != nil {
			return nil, err
		}
		r.addLog("Celular: confirmou " + r.champName(p.ChampionID))
		r.afterPick(p.ChampionID)
	case "import":
		return nil, r.importCiencia()
	case "auto":
		var a rmAuto
		if err := json.Unmarshal(body, &a); err != nil {
			return nil, err
		}
		return r.setAuto(a), nil
	default:
		return nil, errors.New("comando desconhecido")
	}
	return nil, nil
}
