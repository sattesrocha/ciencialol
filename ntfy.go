package main

// Avisos no celular pelo ntfy (https://ntfy.sh): funcionam com o celular bloqueado e até no 4G.
// O PC publica num tópico aleatório só deste PC; o celular assina esse tópico uma vez (QR code).
// Os botões da notificação ("Aceitar", "Escolher Lux") publicam a resposta num SEGUNDO tópico
// aleatório, que o PC fica escutando. Cada resposta leva um código de uso único, que só vale
// pra aquela fila/aquela vez de escolher: uma mensagem vista por outra pessoa não serve pra nada.

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

type rmNtfy struct {
	On     bool   `json:"on"`
	Server string `json:"server"` // padrão https://ntfy.sh
	Topic  string `json:"topic"`  // avisos (o celular assina)
	Reply  string `json:"reply"`  // respostas dos botões (o PC escuta)
	Token  string `json:"token,omitempty"`
}

const ntfyDefaultServer = "https://ntfy.sh"

func randTopic(prefix string) string {
	b := make([]byte, 15)
	_, _ = rand.Read(b)
	return prefix + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
}

func newCode() string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (r *remoteApp) ntfyCfg() rmNtfy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := r.cfg.Ntfy
	if n.Server == "" {
		n.Server = ntfyDefaultServer
	}
	return n
}

// garante tópicos (cria na primeira vez)
func (r *remoteApp) ntfyEnsureTopics() {
	r.mu.Lock()
	if r.cfg.Ntfy.Topic == "" {
		r.cfg.Ntfy.Topic = randTopic("clol-")
	}
	if r.cfg.Ntfy.Reply == "" {
		r.cfg.Ntfy.Reply = randTopic("clolr-")
	}
	r.mu.Unlock()
}

func validServer(s string) (string, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "/")
	if s == "" {
		return ntfyDefaultServer, nil
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.Path != "" {
		return "", errors.New("servidor inválido (use algo como https://ntfy.sh)")
	}
	return u.Scheme + "://" + u.Host, nil
}

type ntfyAction struct {
	Action  string            `json:"action"`
	Label   string            `json:"label"`
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Clear   bool              `json:"clear,omitempty"`
}

type ntfyMsg struct {
	Topic    string       `json:"topic"`
	Title    string       `json:"title,omitempty"`
	Message  string       `json:"message"`
	Priority int          `json:"priority,omitempty"`
	Tags     []string     `json:"tags,omitempty"`
	Actions  []ntfyAction `json:"actions,omitempty"`
	Icon     string       `json:"icon,omitempty"`   // ícone da notificação (logo do Ciência LOL)
	Attach   string       `json:"attach,omitempty"` // imagem grande (arte do campeão)
	Filename string       `json:"filename,omitempty"`
	Click    string       `json:"click,omitempty"` // tocar na notificação abre a página do celular
}

// logo público (site do projeto no GitHub Pages)
const ntfyIcon = "https://sattesrocha.github.io/ciencialol/ui/img/logo.png"

// arte vertical do campeão (Data Dragon, sem versão)
func champArt(alias string) string {
	if alias == "" {
		return ""
	}
	return "https://ddragon.leagueoflegends.com/cdn/img/champion/loading/" + url.PathEscape(alias) + "_0.jpg"
}

var ntfyClient = &http.Client{Timeout: 10 * time.Second}

func (r *remoteApp) ntfySend(m ntfyMsg) error {
	n := r.ntfyCfg()
	if n.Topic == "" {
		return errors.New("avisos no celular desligados")
	}
	m.Topic = n.Topic
	if m.Icon == "" {
		m.Icon = ntfyIcon
	}
	if m.Click == "" {
		if o := r.openBtn(); o != nil {
			m.Click = o.URL
		}
	}
	b, _ := json.Marshal(m)
	req, _ := http.NewRequest("POST", n.Server+"/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}
	resp, err := ntfyClient.Do(req)
	if err != nil {
		return fmt.Errorf("sem conexão com o ntfy (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("o ntfy recusou (%d)", resp.StatusCode)
	}
	return nil
}

// aviso em segundo plano (só se os avisos estiverem ligados)
func (r *remoteApp) notify(m ntfyMsg) {
	if !r.ntfyCfg().On {
		return
	}
	go func() {
		if err := r.ntfySend(m); err != nil {
			r.addLog("Aviso no celular: " + err.Error())
		}
	}()
}

// botão que responde pelo tópico de respostas (funciona no Wi-Fi e no 4G)
func (r *remoteApp) replyBtn(label, body string) ntfyAction {
	n := r.ntfyCfg()
	a := ntfyAction{Action: "http", Label: label, URL: n.Server + "/" + n.Reply, Method: "POST", Body: body, Clear: true}
	if n.Token != "" {
		a.Headers = map[string]string{"Authorization": "Bearer " + n.Token}
	}
	return a
}

// botão "Abrir" com a página do celular (só funciona no mesmo Wi-Fi/Tailscale)
func (r *remoteApp) openBtn() *ntfyAction {
	r.mu.RLock()
	en, port, tok := r.cfg.Enabled, r.cfg.Port, r.cfg.Token
	r.mu.RUnlock()
	ips := localIPv4s()
	if !en || len(ips) == 0 {
		return nil
	}
	return &ntfyAction{Action: "view", Label: "Abrir", URL: fmt.Sprintf("http://%s:%d/?t=%s", ips[0], port, tok)}
}

// ---- eventos (chamados pelo laço do controle) ----

func (r *remoteApp) onReadyCheck(auto bool) {
	if !r.ntfyCfg().On {
		return
	}
	if auto {
		r.notify(ntfyMsg{Title: "Partida encontrada", Message: "Aceita automaticamente ✅", Priority: 3, Tags: []string{"white_check_mark"}})
		return
	}
	code := newCode()
	r.mu.Lock()
	r.rcCode = code
	r.mu.Unlock()
	acts := []ntfyAction{r.replyBtn("✅ Aceitar", "a:"+code), r.replyBtn("❌ Recusar", "d:"+code)}
	if o := r.openBtn(); o != nil {
		acts = append(acts, *o)
	}
	r.notify(ntfyMsg{Title: "Partida encontrada!", Message: "Aceite antes do tempo acabar.", Priority: 5, Tags: []string{"crossed_swords"}, Actions: acts})
}

func (r *remoteApp) onMyTurn(cs *rmChampSelect, a rmAuto) {
	if !r.ntfyCfg().On || cs.MyAction == nil {
		return
	}
	act := cs.MyAction
	title, msg := "Sua vez de escolher", "Escolha o campeão no LoL ou pelo celular."
	if act.Type == "ban" {
		title, msg = "Sua vez de banir", "Bana um campeão no LoL ou pelo celular."
	}
	if a.AutoSelect {
		msg += fmt.Sprintf(" Se você não agir, o automático age quando faltarem %ds.", a.SecondsLeft)
	}
	acts := []ntfyAction{}
	art := ""
	if act.Type == "pick" {
		if c := r.getCiencia(); c != nil && champAvailable(cs, c.Key) {
			if ch := r.champByID(c.Key); ch != nil {
				art = champArt(ch.Alias)
			}
			code := newCode()
			r.mu.Lock()
			r.pkCode, r.pkAction, r.pkChamp = code, act.ID, c.Key
			r.mu.Unlock()
			acts = append(acts, r.replyBtn("Escolher "+c.Champ, "p:"+code))
			msg = "Sua ciência: " + c.Nome + "."
			if a.AutoSelect {
				msg += fmt.Sprintf(" Se você não agir, o automático escolhe quando faltarem %ds.", a.SecondsLeft)
			}
		}
	}
	if o := r.openBtn(); o != nil {
		acts = append(acts, *o)
	}
	m := ntfyMsg{Title: title, Message: msg, Priority: 5, Tags: []string{"hourglass_flowing_sand"}, Actions: acts}
	if art != "" {
		m.Attach, m.Filename = art, "ciencia.jpg"
	}
	r.notify(m)
}

func champAvailable(cs *rmChampSelect, id int) bool {
	ok := false
	for _, x := range cs.Pickable {
		if x == id {
			ok = true
		}
	}
	for _, x := range cs.Unavail {
		if x == id {
			ok = false
		}
	}
	return ok
}

// ---- respostas dos botões ----

func codeEq(a, b string) bool {
	return a != "" && b != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func (r *remoteApp) handleReply(body string) {
	kind, code, ok := strings.Cut(strings.TrimSpace(body), ":")
	if !ok || code == "" {
		return
	}
	r.mu.Lock()
	rc, pk, pkAct, pkChamp := r.rcCode, r.pkCode, r.pkAction, r.pkChamp
	switch {
	case (kind == "a" || kind == "d") && codeEq(code, rc):
		r.rcCode = "" // uso único
	case kind == "p" && codeEq(code, pk):
		r.pkCode = ""
	default:
		r.mu.Unlock()
		return // código errado, velho ou já usado
	}
	r.mu.Unlock()
	switch kind {
	case "a", "d":
		path, msg := "/lol-matchmaking/v1/ready-check/accept", "Celular (notificação): partida aceita"
		if kind == "d" {
			path, msg = "/lol-matchmaking/v1/ready-check/decline", "Celular (notificação): partida recusada"
		}
		if _, _, err := r.lcuDo("POST", path, nil); err != nil {
			r.addLog("Notificação: " + err.Error())
			r.notify(ntfyMsg{Title: "Não deu", Message: "O LoL não aceitou o pedido: " + err.Error(), Priority: 4})
			return
		}
		r.addLog(msg)
	case "p":
		if err := r.lock(pkAct, pkChamp); err != nil {
			r.addLog("Notificação: " + err.Error())
			r.notify(ntfyMsg{Title: "Não deu pra escolher", Message: err.Error(), Priority: 4})
			return
		}
		r.addLog("Celular (notificação): escolheu " + r.champName(pkChamp))
		m := ntfyMsg{Title: "Escolhido: " + r.champName(pkChamp), Message: "Travado no LoL", Priority: 3, Tags: []string{"white_check_mark"}}
		if ch := r.champByID(pkChamp); ch != nil {
			m.Attach, m.Filename = champArt(ch.Alias), "campeao.jpg"
		}
		r.notify(m)
		r.afterPick(pkChamp)
	}
}

// escuta o tópico de respostas enquanto os avisos estiverem ligados
func (r *remoteApp) ntfyListen(ctx context.Context) {
	since := fmt.Sprint(time.Now().Unix())
	stream := &http.Client{} // conexão longa: sem prazo, cancelada pelo ctx
	backoff := 2 * time.Second
	for ctx.Err() == nil {
		n := r.ntfyCfg()
		req, _ := http.NewRequestWithContext(ctx, "GET", n.Server+"/"+n.Reply+"/json?since="+since, nil)
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
					since = m.ID // ao reconectar, continua daqui
				}
				if m.Event == "message" && len(m.Message) < 200 {
					r.handleReply(m.Message)
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

func (r *remoteApp) ntfyApply() {
	n := r.ntfyCfg()
	r.mu.Lock()
	if r.ntfyStop != nil {
		r.ntfyStop()
		r.ntfyStop = nil
	}
	if n.On && n.Reply != "" {
		ctx, cancel := context.WithCancel(context.Background())
		r.ntfyStop = cancel
		go r.ntfyListen(ctx)
	}
	r.mu.Unlock()
	r.ensureLoop()
}

// ---- configuração (app) ----

type rmNtfyInfo struct {
	On         bool   `json:"on"`
	Server     string `json:"server"`
	Topic      string `json:"topic"`
	HasToken   bool   `json:"hasToken"`
	Link       string `json:"link"`       // ntfy://… (assinar com um toque no Android)
	NeedServer bool   `json:"needServer"` // o QR precisa do controle pelo celular ligado
	QR         string `json:"qr"`
}

func (r *remoteApp) ntfyInfo() rmNtfyInfo {
	n := r.ntfyCfg()
	in := rmNtfyInfo{On: n.On, Server: n.Server, Topic: n.Topic, HasToken: n.Token != ""}
	if n.On && n.Topic != "" {
		u, _ := url.Parse(n.Server)
		in.Link = "ntfy://" + u.Host + "/" + n.Topic + "?display=" + url.QueryEscape("Ciência LOL")
		if u.Scheme == "http" {
			in.Link += "&secure=false"
		}
		// a câmera não abre links ntfy:// e transforma texto puro em pesquisa (iPhone): o QR abre a
		// página do celular já na parte do ntfy, com o jeito certo de assinar em cada aparelho
		r.mu.RLock()
		en, port, tok := r.cfg.Enabled, r.cfg.Port, r.cfg.Token
		r.mu.RUnlock()
		ips := localIPv4s()
		in.NeedServer = !en || len(ips) == 0
		if !in.NeedServer {
			if png, err := qrcode.Encode(fmt.Sprintf("http://%s:%d/?t=%s&go=ntfy", ips[0], port, tok), qrcode.Medium, 320); err == nil {
				in.QR = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
			}
		}
	}
	return in
}

type rmNtfySet struct {
	On       bool    `json:"on"`
	Server   *string `json:"server,omitempty"`
	Token    *string `json:"token,omitempty"`
	NewTopic bool    `json:"newTopic,omitempty"`
}

func (r *remoteApp) setNtfy(s rmNtfySet) (rmInfo, error) {
	r.mu.Lock()
	if s.Server != nil {
		v, err := validServer(*s.Server)
		if err != nil {
			r.mu.Unlock()
			return r.info(), err
		}
		r.cfg.Ntfy.Server = v
	}
	if s.Token != nil {
		r.cfg.Ntfy.Token = strings.TrimSpace(*s.Token)
	}
	if s.NewTopic {
		r.cfg.Ntfy.Topic, r.cfg.Ntfy.Reply = "", ""
	}
	r.cfg.Ntfy.On = s.On
	r.mu.Unlock()
	r.ntfyEnsureTopics()
	r.save()
	r.ntfyApply()
	if s.NewTopic {
		r.addLog("Endereço dos avisos trocado: leia o QR code de novo no celular")
	}
	return r.info(), nil
}

func (r *remoteApp) ntfyTest() error {
	if !r.ntfyCfg().On {
		return errors.New("ligue os avisos primeiro")
	}
	return r.ntfySend(ntfyMsg{Title: "Ciência LOL 🧪", Message: "Avisos funcionando! Quando a fila estourar, o botão Aceitar aparece aqui.", Priority: 4, Tags: []string{"tada"}})
}
