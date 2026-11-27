package main

// Controle pelo celular (veio do "LoL Remote"): um servidor na rede local mostra uma página pro
// celular aceitar a partida e banir/escolher na seleção de campeões. Começa DESLIGADO; quando ligado,
// só atende aparelhos da rede local ou do Tailscale e exige o token do link (vira cookie).
// O modo automático (aceitar sozinho, banir/escolher se faltar pouco tempo) roda aqui mesmo,
// usando a mesma conexão com o LoL do resto do app. Com a ciência sorteada, o campeão dela vem
// primeiro na lista e a build pode ir pro LoL sozinha ao travar.

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

//go:embed remote/index.html remote/logo.png remote/cinzel.ttf
var remoteFS embed.FS

// ---- tipos da LCU usados aqui ----

type rmChamp struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type rmAction struct {
	ID           int    `json:"id"`
	ActorCellID  int    `json:"actorCellId"`
	ChampionID   int    `json:"championId"`
	Completed    bool   `json:"completed"`
	IsInProgress bool   `json:"isInProgress"`
	Type         string `json:"type"` // "ban", "pick", "ten_bans_reveal"
}

type rmPlayer struct {
	CellID             int    `json:"cellId"`
	ChampionID         int    `json:"championId"`
	ChampionPickIntent int    `json:"championPickIntent"`
	AssignedPosition   string `json:"assignedPosition"`
}

type rmSession struct {
	Actions           [][]rmAction `json:"actions"`
	LocalPlayerCellID int          `json:"localPlayerCellId"`
	MyTeam            []rmPlayer   `json:"myTeam"`
	TheirTeam         []rmPlayer   `json:"theirTeam"`
	Timer             struct {
		AdjustedTimeLeftInPhase float64 `json:"adjustedTimeLeftInPhase"` // quanto faltava…
		InternalNowInEpochMs    float64 `json:"internalNowInEpochMs"`    // …neste instante (o LoL não atualiza sozinho)
		IsInfinite              bool    `json:"isInfinite"`              // partida personalizada: sem tempo
		Phase                   string  `json:"phase"`
	} `json:"timer"`
}

type rmReady struct {
	State          string  `json:"state"`
	PlayerResponse string  `json:"playerResponse"`
	Timer          float64 `json:"timer"`
}

// ---- estado mostrado no celular e no app ----

type rmMyAction struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	ChampionID int    `json:"championId"`
}

type rmSlot struct {
	ChampionID int    `json:"championId"`
	Intent     int    `json:"intent"`
	Position   string `json:"position"`
	Me         bool   `json:"me"`
}

type rmChampSelect struct {
	Phase     string      `json:"phase"`
	TimeLeft  float64     `json:"timeLeft"`
	Infinite  bool        `json:"infinite"`
	MyAction  *rmMyAction `json:"myAction"`
	MyChamp   int         `json:"myChamp"`
	MyTeam    []rmSlot    `json:"myTeam"`
	TheirTeam []rmSlot    `json:"theirTeam"`
	Bans      []int       `json:"bans"`
	Pickable  []int       `json:"pickable"`
	Bannable  []int       `json:"bannable"`
	Unavail   []int       `json:"unavailable"`
}

// ciência sorteada pra este PC (o app manda quando muda o sorteio)
type rmCiencia struct {
	Key     int    `json:"key"`     // id numérico do campeão (mesmo da LCU)
	Champ   string `json:"champ"`   // nome do campeão
	Nome    string `json:"nome"`    // nome da ciência
	Payload string `json:"payload"` // build pronta pro importar (vazio = sem build)
}

type rmView struct {
	Connected   bool           `json:"connected"`
	Error       string         `json:"error,omitempty"`
	Phase       string         `json:"phase"`
	ReadyCheck  *rmReady       `json:"readyCheck,omitempty"`
	ChampSelect *rmChampSelect `json:"champSelect,omitempty"`
	Auto        rmAuto         `json:"auto"`
	Ciencia     *rmCiencia     `json:"ciencia,omitempty"`
	Imported    bool           `json:"imported"`           // build da ciência já foi pro LoL nesta seleção
	NtfyLink    string         `json:"ntfyLink,omitempty"` // ntfy://… pra assinar com um toque (página do celular)
	NtfyTopic   string         `json:"ntfyTopic,omitempty"`
	Stable      string         `json:"stable,omitempty"` // link fixo pro atalho (só pra quem já entrou)
	Log         []string       `json:"log"`
}

type rmAuto struct {
	AutoAccept  bool  `json:"autoAccept"`
	AutoSelect  bool  `json:"autoSelect"`
	UseCiencia  bool  `json:"useCiencia"` // campeão da ciência sorteada vem primeiro
	AutoImport  bool  `json:"autoImport"` // ao travar o campeão da ciência, importa a build
	BanList     []int `json:"banList"`
	PickList    []int `json:"pickList"`
	SecondsLeft int   `json:"secondsLeft"`
}

type rmConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Token   string `json:"token"`
	// caixa postal do endereço (atalho fixo do celular)
	AddrTopic string `json:"addrTopic"`
	Auto      rmAuto `json:"auto"`
	Ntfy      rmNtfy `json:"ntfy"`
}

type remoteApp struct {
	mu        sync.RWMutex
	retired   bool // outra janela do Ciência LOL assumiu: esta para de agir e de avisar
	path      string
	cfg       rmConfig
	conn      *lcuConn
	view      rmView
	champs    []rmChamp // do LoL (quando aberto)
	ddVer     string    // versão do Data Dragon (o app manda): nomes e ícones mesmo com o LoL fechado
	ddChamps  []rmChamp
	logs      []string
	acted     map[int]time.Time
	lastPhase string
	ciencia   *rmCiencia
	imported  bool
	srv       *http.Server
	loopOn    bool
	fails     map[string][]time.Time // tentativas com token errado por IP
	// avisos pelo ntfy
	ntfyStop          context.CancelFunc
	addrStop          context.CancelFunc
	rcCode, pkCode    string // códigos de uso único dos botões da notificação
	pkAction, pkChamp int
	rcSeen            bool // já avisou desta fila
	lastTurn          int  // última vez de banir/escolher avisada
}

var remote = &remoteApp{acted: map[int]time.Time{}, fails: map[string][]time.Time{}}

func remoteInit(dir string) {
	r := remote
	r.path = filepath.Join(dir, "remote.json")
	r.cfg = rmConfig{Port: 8787, Auto: rmAuto{SecondsLeft: 8, UseCiencia: true, AutoImport: true}}
	if b, err := os.ReadFile(r.path); err == nil {
		_ = json.Unmarshal(b, &r.cfg)
	}
	if r.cfg.Port < 1024 || r.cfg.Port > 65535 {
		r.cfg.Port = 8787
	}
	if len(r.cfg.Token) < 32 {
		r.cfg.Token = newRemoteToken()
	}
	if r.cfg.AddrTopic == "" {
		r.cfg.AddrTopic = randTopic("clola-")
	}
	r.save()
	if r.cfg.Ntfy.On {
		r.ntfyEnsureTopics()
		r.save()
		r.ntfyApply()
	}
	if r.cfg.Enabled {
		if err := r.startServer(); err != nil {
			r.addLog("Não deu pra ligar o controle pelo celular: " + err.Error())
		}
	}
	r.ensureLoop()
}

func newRemoteToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *remoteApp) save() {
	r.mu.RLock()
	b, _ := json.MarshalIndent(r.cfg, "", "  ")
	r.mu.RUnlock()
	if r.path != "" {
		_ = os.WriteFile(r.path, b, 0o600)
	}
}

func (r *remoteApp) addLog(s string) {
	line := time.Now().Format("15:04:05") + " " + s
	r.mu.Lock()
	r.logs = append(r.logs, line)
	if len(r.logs) > 40 {
		r.logs = r.logs[len(r.logs)-40:]
	}
	r.mu.Unlock()
}

func (r *remoteApp) autoCfg() rmAuto {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Auto
}

func (r *remoteApp) setAuto(a rmAuto) rmAuto {
	if a.SecondsLeft < 1 {
		a.SecondsLeft = 1
	}
	if a.SecondsLeft > 60 {
		a.SecondsLeft = 60
	}
	clip := func(l []int) []int {
		out := []int{}
		seen := map[int]bool{}
		for _, id := range l {
			if id > 0 && !seen[id] && len(out) < 30 {
				seen[id] = true
				out = append(out, id)
			}
		}
		return out
	}
	a.BanList, a.PickList = clip(a.BanList), clip(a.PickList)
	r.mu.Lock()
	r.cfg.Auto = a
	r.mu.Unlock()
	r.save()
	r.ensureLoop()
	return a
}

func (r *remoteApp) setCiencia(c *rmCiencia) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c != nil && c.Key <= 0 {
		c = nil
	}
	if c == nil || r.ciencia == nil || c.Key != r.ciencia.Key || c.Payload != r.ciencia.Payload {
		r.imported = false
	}
	r.ciencia = c
}

func (r *remoteApp) getCiencia() *rmCiencia {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.ciencia == nil {
		return nil
	}
	c := *r.ciencia
	return &c
}

// ---- conversa com o LoL (conexão guardada; se cair, reconecta no próximo pedido) ----

func (r *remoteApp) lcuDo(method, path string, body any) ([]byte, int, error) {
	for attempt := 0; attempt < 2; attempt++ {
		r.mu.Lock()
		c := r.conn
		r.mu.Unlock()
		if c == nil {
			nc, err := lcuConnectLog(func(string, ...any) {})
			if err != nil {
				return nil, 0, err
			}
			nc.cl.Timeout = 6 * time.Second
			r.mu.Lock()
			r.conn = nc
			r.mu.Unlock()
			c = nc
		}
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, c.base+path, rd)
		req.SetBasicAuth("riot", c.pass)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := c.cl.Do(req)
		if err != nil { // o LoL pode ter reiniciado e trocado de porta
			r.mu.Lock()
			if r.conn == c {
				r.conn = nil
			}
			r.mu.Unlock()
			continue
		}
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			r.mu.Lock()
			r.conn = nil
			r.mu.Unlock()
			continue
		}
		if resp.StatusCode >= 400 {
			msg := strings.TrimSpace(string(data))
			var e struct {
				Message string `json:"message"`
			}
			if json.Unmarshal(data, &e) == nil && e.Message != "" {
				msg = e.Message
			}
			if len(msg) > 160 {
				msg = msg[:160]
			}
			return data, resp.StatusCode, fmt.Errorf("o LoL recusou (%d): %s", resp.StatusCode, msg)
		}
		return data, resp.StatusCode, nil
	}
	return nil, 0, errors.New("não achei o LoL aberto neste PC")
}

func (r *remoteApp) getJSON(path string, v any) error {
	data, _, err := r.lcuDo("GET", path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ---- laço: roda enquanto o controle pelo celular ou algum automático estiver ligado ----

func (r *remoteApp) active() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a := r.cfg.Auto
	return r.cfg.Enabled || a.AutoAccept || a.AutoSelect || r.cfg.Ntfy.On
}

func (r *remoteApp) ensureLoop() {
	r.mu.Lock()
	if r.loopOn {
		r.mu.Unlock()
		return
	}
	r.loopOn = true
	r.mu.Unlock()
	go func() {
		for {
			if !r.active() {
				r.mu.Lock()
				r.loopOn = false
				r.view = rmView{Phase: "Desligado"}
				r.conn = nil
				r.mu.Unlock()
				return
			}
			r.tick()
			time.Sleep(900 * time.Millisecond)
		}
	}()
}

func (r *remoteApp) tick() {
	r.mu.RLock()
	retired := r.retired
	r.mu.RUnlock()
	if retired {
		return
	}
	var phase string
	if err := r.getJSON("/lol-gameflow/v1/gameflow-phase", &phase); err != nil {
		r.mu.Lock()
		r.view = rmView{Connected: false, Error: err.Error(), Phase: "Offline"}
		r.mu.Unlock()
		return
	}
	r.mu.RLock()
	needChamps := len(r.champs) == 0
	r.mu.RUnlock()
	if needChamps {
		var cs []rmChamp
		if err := r.getJSON("/lol-game-data/assets/v1/champion-summary.json", &cs); err == nil {
			out := cs[:0]
			for _, c := range cs {
				if c.ID > 0 {
					out = append(out, c)
				}
			}
			sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
			r.mu.Lock()
			r.champs = out
			r.mu.Unlock()
		}
	}
	if phase != r.lastPhase {
		r.addLog("Fase: " + phaseName(phase))
		r.mu.Lock()
		if phase != "ChampSelect" {
			r.acted = map[int]time.Time{}
			r.imported = false
			r.pkCode, r.lastTurn = "", 0
		}
		if phase != "ReadyCheck" {
			r.rcCode, r.rcSeen = "", false // a fila acabou: o botão da notificação deixa de valer
		}
		r.mu.Unlock()
		r.lastPhase = phase
	}

	v := rmView{Connected: true, Phase: phase}
	a := r.autoCfg()
	switch phase {
	case "ReadyCheck":
		var rc rmReady
		if err := r.getJSON("/lol-matchmaking/v1/ready-check", &rc); err == nil {
			v.ReadyCheck = &rc
			r.mu.Lock()
			first := !r.rcSeen && rc.PlayerResponse == "None" && rc.State == "InProgress"
			if first {
				r.rcSeen = true
			}
			r.mu.Unlock()
			autoOK := false
			if a.AutoAccept && rc.PlayerResponse == "None" && rc.State == "InProgress" {
				if _, _, err := r.lcuDo("POST", "/lol-matchmaking/v1/ready-check/accept", nil); err == nil {
					r.addLog("Automático: partida aceita")
					autoOK = true
				}
			}
			if first {
				r.onReadyCheck(autoOK)
			}
		}
	case "ChampSelect":
		if cs, err := r.buildChampSelect(); err == nil {
			v.ChampSelect = cs
			if cs.MyAction != nil {
				r.mu.Lock()
				newTurn := cs.MyAction.ID != r.lastTurn
				r.lastTurn = cs.MyAction.ID
				r.mu.Unlock()
				if newTurn {
					r.onMyTurn(cs, a)
					if a.AutoSelect {
						r.explainAuto(cs, a)
					}
				}
			}
			if a.AutoSelect {
				r.autoSelect(cs, a)
			}
		} else {
			v.Error = err.Error()
		}
	}
	r.mu.Lock()
	r.view = v
	r.mu.Unlock()
}

func phaseName(p string) string {
	m := map[string]string{"None": "no menu", "Lobby": "na sala", "Matchmaking": "na fila", "ReadyCheck": "partida encontrada",
		"ChampSelect": "seleção de campeões", "GameStart": "carregando", "InProgress": "em jogo", "WaitingForStats": "fim de jogo",
		"PreEndOfGame": "fim de jogo", "EndOfGame": "fim de jogo", "Reconnect": "reconectar"}
	if n, ok := m[p]; ok {
		return n
	}
	return p
}

func (r *remoteApp) buildChampSelect() (*rmChampSelect, error) {
	var s rmSession
	if err := r.getJSON("/lol-champ-select/v1/session", &s); err != nil {
		return nil, err
	}
	cs := &rmChampSelect{Phase: s.Timer.Phase, TimeLeft: timeLeftNow(s.Timer.AdjustedTimeLeftInPhase, s.Timer.InternalNowInEpochMs, time.Now()), Infinite: s.Timer.IsInfinite}
	unavail := map[int]bool{}
	for _, p := range s.MyTeam {
		me := p.CellID == s.LocalPlayerCellID
		if me {
			cs.MyChamp = p.ChampionID
		}
		cs.MyTeam = append(cs.MyTeam, rmSlot{p.ChampionID, p.ChampionPickIntent, p.AssignedPosition, me})
		if p.ChampionID > 0 && !me {
			unavail[p.ChampionID] = true
		}
	}
	for _, p := range s.TheirTeam {
		cs.TheirTeam = append(cs.TheirTeam, rmSlot{ChampionID: p.ChampionID, Position: p.AssignedPosition})
		if p.ChampionID > 0 {
			unavail[p.ChampionID] = true
		}
	}
	for _, group := range s.Actions {
		for _, act := range group {
			if act.Completed && act.ChampionID > 0 {
				unavail[act.ChampionID] = true
				if act.Type == "ban" {
					cs.Bans = append(cs.Bans, act.ChampionID)
				}
			}
			if act.ActorCellID == s.LocalPlayerCellID && act.IsInProgress && !act.Completed && (act.Type == "ban" || act.Type == "pick") {
				cs.MyAction = &rmMyAction{ID: act.ID, Type: act.Type, ChampionID: act.ChampionID}
			}
		}
	}
	_ = r.getJSON("/lol-champ-select/v1/pickable-champion-ids", &cs.Pickable)
	_ = r.getJSON("/lol-champ-select/v1/bannable-champion-ids", &cs.Bannable)
	for id := range unavail {
		cs.Unavail = append(cs.Unavail, id)
	}
	return cs, nil
}

// lista efetiva de picks: campeão da ciência primeiro (se ligado), depois a lista de prioridade
func (r *remoteApp) pickOrder(a rmAuto) []int {
	out := []int{}
	if c := r.getCiencia(); a.UseCiencia && c != nil && c.Key > 0 {
		out = append(out, c.Key)
	}
	for _, id := range a.PickList {
		if len(out) == 0 || id != out[0] {
			out = append(out, id)
		}
	}
	return out
}

// O LoL manda "faltavam X ms no instante T" e só atualiza quando algo muda na seleção.
// O tempo que falta AGORA é X - (agora - T). Sem T (ou com relógio muito torto), usa X.
func timeLeftNow(adjustedMs, internalNowMs float64, now time.Time) float64 {
	left := adjustedMs
	if internalNowMs > 0 {
		el := float64(now.UnixMilli()) - internalNowMs
		if el >= 0 && el < 10*60*1000 {
			left = adjustedMs - el
		}
	}
	if left < 0 {
		left = 0
	}
	return left / 1000
}

func (r *remoteApp) autoSelect(cs *rmChampSelect, a rmAuto) {
	act := cs.MyAction
	if act == nil || cs.Infinite || cs.TimeLeft > float64(a.SecondsLeft) {
		return
	}
	r.mu.Lock() // no máximo uma tentativa a cada 2 s por ação
	if time.Since(r.acted[act.ID]) < 2*time.Second {
		r.mu.Unlock()
		return
	}
	r.acted[act.ID] = time.Now()
	r.mu.Unlock()
	list, allowed := r.pickOrder(a), cs.Pickable
	if act.Type == "ban" {
		list, allowed = a.BanList, cs.Bannable
	}
	ok := map[int]bool{}
	for _, id := range allowed {
		ok[id] = true
	}
	for _, id := range cs.Unavail {
		delete(ok, id)
	}
	if act.Type == "ban" { // nunca bane o campeão da própria ciência
		if c := r.getCiencia(); c != nil {
			delete(ok, c.Key)
		}
	}
	for _, id := range list {
		if ok[id] {
			if err := r.lock(act.ID, id); err != nil {
				r.addLog("Automático falhou: " + err.Error())
				continue
			}
			verb := "escolheu"
			if act.Type == "ban" {
				verb = "baniu"
			}
			r.addLog("Automático: " + verb + " " + r.champName(id))
			r.notify(ntfyMsg{Title: "Automático " + verb + " " + r.champName(id), Message: "Feito no LoL ✅", Priority: 3, Tags: []string{"robot"}})
			if act.Type == "pick" {
				r.afterPick(id)
			}
			return
		}
	}
	if act.Type == "ban" {
		r.addLog("Automático: nenhum campeão da lista de bans está disponível")
	} else {
		r.addLog("Automático: nenhum campeão da lista de picks está disponível")
	}
}

// no começo da vez: avisa no registro quando o automático não vai poder agir
func (r *remoteApp) explainAuto(cs *rmChampSelect, a rmAuto) {
	act := cs.MyAction
	what := "escolher"
	if act.Type == "ban" {
		what = "banir"
	}
	switch {
	case cs.Infinite:
		r.addLog("Automático: esta seleção não tem tempo (partida personalizada), então ele não vai " + what)
	case act.Type == "ban" && len(a.BanList) == 0:
		r.addLog("Automático: sua lista de bans está vazia, então ele não vai banir")
	case act.Type == "pick" && len(r.pickOrder(a)) == 0:
		r.addLog("Automático: sem ciência sorteada e com a lista de picks vazia, então ele não vai escolher")
	default:
		r.addLog(fmt.Sprintf("Automático: vai %s quando faltarem %ds, se você não agir", what, a.SecondsLeft))
	}
}

func (r *remoteApp) champName(id int) string {
	if c := r.champByID(id); c != nil {
		return c.Name
	}
	return "#" + strconv.Itoa(id)
}

func (r *remoteApp) champByID(id int) *rmChamp {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, l := range [][]rmChamp{r.champs, r.ddChamps} {
		for _, c := range l {
			if c.ID == id {
				c := c
				return &c
			}
		}
	}
	return nil
}

// o app manda a lista do Data Dragon (nomes e ícones mesmo com o LoL fechado)
func (r *remoteApp) setDDChamps(ver string, list []rmChamp) {
	out := []rmChamp{}
	for _, c := range list {
		if c.ID > 0 && c.Name != "" && len(out) < 400 {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	r.mu.Lock()
	r.ddVer, r.ddChamps = ver, out
	r.mu.Unlock()
}

func (r *remoteApp) hover(actionID, champID int) error {
	if actionID <= 0 || champID <= 0 {
		return errors.New("escolha inválida")
	}
	_, _, err := r.lcuDo("PATCH", "/lol-champ-select/v1/session/actions/"+strconv.Itoa(actionID), map[string]any{"championId": champID})
	return err
}

func (r *remoteApp) lock(actionID, champID int) error {
	if err := r.hover(actionID, champID); err != nil {
		return err
	}
	_, _, err := r.lcuDo("POST", "/lol-champ-select/v1/session/actions/"+strconv.Itoa(actionID)+"/complete", nil)
	return err
}

// travou o campeão da ciência: importa a build sozinho (se ligado)
func (r *remoteApp) afterPick(champID int) {
	c := r.getCiencia()
	if c == nil || c.Key != champID || c.Payload == "" || !r.autoCfg().AutoImport {
		return
	}
	go func() {
		if err := r.importCiencia(); err != nil {
			r.addLog("Build da ciência: " + err.Error())
		}
	}()
}

func (r *remoteApp) importCiencia() error {
	c := r.getCiencia()
	if c == nil || c.Payload == "" {
		return errors.New("não tem build da ciência pra importar")
	}
	res, err := lcuImport(c.Payload, false)
	if err != nil {
		return err
	}
	if strings.HasPrefix(res.Runes, "full:") {
		res.Runes = "páginas de runas cheias (importe pelo app pra trocar uma)"
	}
	r.mu.Lock()
	r.imported = true
	r.mu.Unlock()
	line := fmt.Sprintf("runas: %s · feitiços: %s · itens: %s", res.Runes, res.Spells, res.Items)
	r.addLog("Build de " + c.Nome + " no LoL — " + line)
	r.notify(ntfyMsg{Title: "Build de " + c.Nome + " no LoL ⚡", Message: line, Priority: 2})
	return nil
}

func (r *remoteApp) viewNow() rmView {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v := r.view
	v.Auto = r.cfg.Auto
	if r.ciencia != nil {
		c := *r.ciencia
		c.Payload = map[bool]string{true: "1", false: ""}[c.Payload != ""] // o celular só precisa saber se tem build
		v.Ciencia = &c
	}
	v.Imported = r.imported
	if r.cfg.Enabled && r.cfg.AddrTopic != "" { // montado aqui (já com a trava): não chama stableLink
		v.Stable = stableBase + "#" + r.cfg.AddrTopic + "." + r.cfg.Token
		if r.cfg.Ntfy.Server != "" && r.cfg.Ntfy.Server != ntfyDefaultServer {
			v.Stable += "&s=" + url.QueryEscape(r.cfg.Ntfy.Server)
		}
		v.Stable += "&a" // botão "Criar atalho" da página do celular
	}
	if r.cfg.Ntfy.On && r.cfg.Ntfy.Topic != "" {
		srv := r.cfg.Ntfy.Server
		if srv == "" {
			srv = ntfyDefaultServer
		}
		if u, err := url.Parse(srv); err == nil {
			v.NtfyLink = "ntfy://" + u.Host + "/" + r.cfg.Ntfy.Topic + "?display=" + url.QueryEscape("Ciência LOL")
			if u.Scheme == "http" {
				v.NtfyLink += "&secure=false"
			}
		}
		v.NtfyTopic = r.cfg.Ntfy.Topic
	}
	v.Log = append([]string(nil), r.logs...)
	return v
}

// ---- servidor pro celular ----

func (r *remoteApp) startServer() error {
	r.mu.Lock()
	if r.srv != nil {
		r.mu.Unlock()
		return nil
	}
	port := r.cfg.Port
	r.mu.Unlock()
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil && r.askHandoff(port) { // um Ciência LOL antigo ainda aberto: ele solta a porta pra este
		for i := 0; i < 20 && err != nil; i++ {
			time.Sleep(250 * time.Millisecond)
			ln, err = net.Listen("tcp", ":"+strconv.Itoa(port))
		}
	}
	if err != nil {
		if r.isOldCiencia(port) {
			return fmt.Errorf("outro Ciência LOL aberto neste PC está usando o controle do celular. Feche as outras janelas do Ciência LOL")
		}
		return fmt.Errorf("a porta %d está ocupada por outro programa", port)
	}
	srv := &http.Server{Handler: r.handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second}
	r.addrEnsureTopic()
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.srv = srv
	r.addrStop = cancel
	r.mu.Unlock()
	go func() { _ = srv.Serve(ln) }()
	go r.addrLoop(ctx)
	go r.relayLoop(ctx)
	r.addLog("Controle pelo celular ligado na porta " + strconv.Itoa(port))
	return nil
}

func (r *remoteApp) stopServer() {
	r.mu.Lock()
	srv := r.srv
	r.srv = nil
	if r.addrStop != nil {
		r.addrStop()
		r.addrStop = nil
	}
	r.mu.Unlock()
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		r.addLog("Controle pelo celular desligado")
	}
}

// pede pro Ciência LOL que já está na porta (mesma chave, mesmo PC) soltar o controle
func (r *remoteApp) askHandoff(port int) bool {
	req, _ := http.NewRequest("POST", "http://127.0.0.1:"+strconv.Itoa(port)+"/api/handoff", nil)
	req.Header.Set("X-Token", r.token())
	resp, err := (&http.Client{Timeout: 2 * time.Second}).Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// a porta está com uma versão antiga do Ciência LOL (que não sabe soltar)?
func (r *remoteApp) isOldCiencia(port int) bool {
	resp, err := (&http.Client{Timeout: 2 * time.Second}).Get("http://127.0.0.1:" + strconv.Itoa(port) + "/manifest.webmanifest")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
	return strings.Contains(string(b), "Ciência LOL")
}

// só aparelhos da rede local, do Tailscale (100.64.0.0/10) ou deste PC
func allowedRemoteIP(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
		if ip4[0] == 100 && ip4[1]&0xC0 == 64 {
			return true
		}
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

func (r *remoteApp) token() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Token
}

func tokEq(a, b string) bool {
	return len(a) > 0 && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// muitas tentativas com token errado do mesmo IP: bloqueia por um minuto
func (r *remoteApp) tooManyFails(ip string, fail bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	l := r.fails[ip][:0]
	for _, t := range r.fails[ip] {
		if now.Sub(t) < time.Minute {
			l = append(l, t)
		}
	}
	if fail {
		l = append(l, now)
	}
	r.fails[ip] = l
	return len(l) > 10
}

func (r *remoteApp) authed(req *http.Request) bool {
	tok := r.token()
	if c, err := req.Cookie("clr"); err == nil && tokEq(c.Value, tok) {
		return true
	}
	return tokEq(req.Header.Get("X-Token"), tok)
}

func (r *remoteApp) handler() http.Handler {
	mux := http.NewServeMux()
	jsonOut := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		ip, _, _ := net.SplitHostPort(req.RemoteAddr)
		// Primeiro acesso pelo link: grava o cookie e já mostra a página (sem redirecionar). Um redirecionamento
		// aqui fazia o iPhone não mandar o cookie quando o link veio do atalho fixo (site → PC).
		t := req.URL.Query().Get("t")
		if t != "" {
			if !tokEq(t, r.token()) {
				r.tooManyFails(ip, true)
				http.Error(w, "Link inválido ou antigo. Abra o link novo mostrado no Ciência LOL.", http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "clr", Value: t, Path: "/", MaxAge: 60 * 60 * 24 * 365, HttpOnly: true, SameSite: http.SameSiteLaxMode})
		} else if !r.authed(req) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Abra o link (ou o QR code) mostrado no Ciência LOL, na aba Celular.")
			return
		}
		b, _ := remoteFS.ReadFile("remote/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https://ddragon.leagueoflegends.com; font-src 'self'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; connect-src 'self'; manifest-src 'self'")
		w.Write(b)
	})
	static := func(path, file, ctype string) {
		mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			b, _ := remoteFS.ReadFile("remote/" + file)
			w.Header().Set("Content-Type", ctype)
			w.Header().Set("Cache-Control", "max-age=604800")
			w.Write(b)
		})
	}
	static("/logo.png", "logo.png", "image/png")
	static("/cinzel.ttf", "cinzel.ttf", "font/ttf")
	mux.HandleFunc("/manifest.webmanifest", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		fmt.Fprint(w, `{"name":"Ciência LOL","short_name":"Ciência LOL","start_url":"/","scope":"/","display":"standalone","orientation":"portrait","background_color":"#070b14","theme_color":"#070b14","icons":[{"src":"/logo.png","sizes":"320x320","type":"image/png","purpose":"any maskable"}]}`)
	})
	api := func(path string, post bool, h func(w http.ResponseWriter, req *http.Request) (any, error)) {
		mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			if !r.authed(req) {
				http.Error(w, "não autorizado", http.StatusUnauthorized)
				return
			}
			if post {
				if req.Method != http.MethodPost {
					http.Error(w, "use POST", http.StatusMethodNotAllowed)
					return
				}
				if o := req.Header.Get("Origin"); o != "" { // pedido vindo de outro site: recusa
					if u, err := url.Parse(o); err != nil || u.Host != req.Host {
						http.Error(w, "origem inválida", http.StatusForbidden)
						return
					}
				}
			}
			v, err := h(w, req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			if v != nil {
				jsonOut(w, v)
			}
		})
	}
	ok := map[string]bool{"ok": true}
	// outra janela do Ciência LOL (versão nova) neste PC pediu o controle do celular: solta a porta
	api("/api/handoff", true, func(w http.ResponseWriter, req *http.Request) (any, error) {
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
			return nil, fmt.Errorf("só deste PC")
		}
		go func() {
			time.Sleep(200 * time.Millisecond)
			r.mu.Lock()
			r.retired = true
			if r.ntfyStop != nil {
				r.ntfyStop()
				r.ntfyStop = nil
			}
			r.mu.Unlock()
			r.stopServer()
			r.addLog("Outra janela do Ciência LOL assumiu o controle pelo celular")
		}()
		return ok, nil
	})
	type pickBody struct{ ActionID, ChampionID int }
	readPick := func(req *http.Request) (pickBody, error) {
		var b pickBody
		err := json.NewDecoder(io.LimitReader(req.Body, 4096)).Decode(&b)
		return b, err
	}
	api("/api/state", false, func(w http.ResponseWriter, req *http.Request) (any, error) { return r.viewNow(), nil })
	api("/api/champions", false, func(w http.ResponseWriter, req *http.Request) (any, error) {
		r.mu.RLock()
		defer r.mu.RUnlock()
		list := r.champs
		if len(list) == 0 {
			list = r.ddChamps
		}
		return map[string]any{"ver": r.ddVer, "list": list}, nil
	})
	api("/api/accept", true, func(w http.ResponseWriter, req *http.Request) (any, error) {
		if _, _, err := r.lcuDo("POST", "/lol-matchmaking/v1/ready-check/accept", nil); err != nil {
			return nil, err
		}
		r.addLog("Celular: partida aceita")
		return ok, nil
	})
	api("/api/decline", true, func(w http.ResponseWriter, req *http.Request) (any, error) {
		if _, _, err := r.lcuDo("POST", "/lol-matchmaking/v1/ready-check/decline", nil); err != nil {
			return nil, err
		}
		r.addLog("Celular: partida recusada")
		return ok, nil
	})
	api("/api/hover", true, func(w http.ResponseWriter, req *http.Request) (any, error) {
		b, err := readPick(req)
		if err != nil {
			return nil, err
		}
		return ok, r.hover(b.ActionID, b.ChampionID)
	})
	api("/api/lock", true, func(w http.ResponseWriter, req *http.Request) (any, error) {
		b, err := readPick(req)
		if err != nil {
			return nil, err
		}
		if err := r.lock(b.ActionID, b.ChampionID); err != nil {
			return nil, err
		}
		r.addLog("Celular: confirmou " + r.champName(b.ChampionID))
		r.afterPick(b.ChampionID)
		return ok, nil
	})
	api("/api/import", true, func(w http.ResponseWriter, req *http.Request) (any, error) {
		return ok, r.importCiencia()
	})
	api("/api/auto", false, func(w http.ResponseWriter, req *http.Request) (any, error) {
		if req.Method == http.MethodPost {
			if o := req.Header.Get("Origin"); o != "" {
				if u, err := url.Parse(o); err != nil || u.Host != req.Host {
					return nil, errors.New("origem inválida")
				}
			}
			var a rmAuto
			if err := json.NewDecoder(io.LimitReader(req.Body, 16384)).Decode(&a); err != nil {
				return nil, err
			}
			return r.setAuto(a), nil
		}
		return r.autoCfg(), nil
	})
	api("/icon/", false, func(w http.ResponseWriter, req *http.Request) (any, error) {
		id, err := strconv.Atoi(strings.TrimPrefix(req.URL.Path, "/icon/"))
		if err != nil || id <= 0 || id > 100000 {
			http.NotFound(w, req)
			return nil, nil
		}
		data, _, err := r.lcuDo("GET", "/lol-game-data/assets/v1/champion-icons/"+strconv.Itoa(id)+".png", nil)
		if err != nil {
			http.NotFound(w, req)
			return nil, nil
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=86400")
		w.Write(data)
		return nil, nil
	})
	// porta de entrada: só rede local/Tailscale e sem martelar token
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !allowedRemoteIP(req.RemoteAddr) {
			http.Error(w, "só aparelhos da mesma rede", http.StatusForbidden)
			return
		}
		ip, _, _ := net.SplitHostPort(req.RemoteAddr)
		if r.tooManyFails(ip, false) {
			http.Error(w, "muitas tentativas; espere um minuto", http.StatusTooManyRequests)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		mux.ServeHTTP(w, req)
	})
}

// ---- endereços e QR code pro app ----

func localIPv4s() []string {
	var priv, tail []string
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || ipn.IP.To4() == nil || ipn.IP.IsLinkLocalUnicast() {
				continue
			}
			ip := ipn.IP.To4()
			if ip[0] == 100 && ip[1]&0xC0 == 64 {
				tail = append(tail, ip.String()) // Tailscale por último
			} else if ip.IsPrivate() {
				priv = append(priv, ip.String())
			}
		}
	}
	return append(priv, tail...)
}

type rmInfo struct {
	Enabled bool       `json:"enabled"`
	Port    int        `json:"port"`
	Links   []string   `json:"links"`
	QR      string     `json:"qr"` // data:image/png do link fixo
	Stable  string     `json:"stable"`
	View    rmView     `json:"view"`
	Ntfy    rmNtfyInfo `json:"ntfy"`
}

func (r *remoteApp) info() rmInfo {
	r.mu.RLock()
	en, port, tok := r.cfg.Enabled, r.cfg.Port, r.cfg.Token
	r.mu.RUnlock()
	in := rmInfo{Enabled: en, Port: port, View: r.viewNow(), Ntfy: r.ntfyInfo()}
	if en {
		for _, ip := range localIPv4s() {
			in.Links = append(in.Links, fmt.Sprintf("http://%s:%d/?t=%s", ip, port, tok))
		}
		if len(in.Links) > 0 {
			in.Stable = r.stableLink("") // QR: abre direto no PC
			if png, err := qrcode.Encode(in.Stable, qrcode.Medium, 320); err == nil {
				in.QR = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
			}
		}
	}
	return in
}

func (r *remoteApp) setEnabled(on bool) (rmInfo, error) {
	if on {
		if err := r.startServer(); err != nil {
			return r.info(), err
		}
	} else {
		r.stopServer()
	}
	r.mu.Lock()
	r.cfg.Enabled = on
	r.mu.Unlock()
	r.save()
	r.ensureLoop()
	return r.info(), nil
}

// gera link novo (o antigo para de funcionar em todos os celulares)
func (r *remoteApp) newToken() rmInfo {
	r.mu.Lock()
	r.cfg.Token = newRemoteToken()
	r.mu.Unlock()
	r.save()
	r.addLog("Link do celular trocado (o anterior não funciona mais)")
	return r.info()
}
