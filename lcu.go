package main

// Importar pro LoL: conversa com o cliente do League of Legends aberto neste PC
// (a "porta local" que o próprio cliente abre, com a senha do arquivo lockfile).
// Só age quando a pessoa clica no botão. Não lê memória nem mexe em nada durante a partida.

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const lcuPageName = "Ciência LOL"

type lcuConn struct {
	base     string // https://127.0.0.1:porta
	pass     string
	cl       *http.Client
	deadline time.Time // se definido: nenhum pedido passa desse horário (o importar nunca fica preso)
	log      bool      // anota cada pedido no lcu.log (usado no importar)
}

// lockfile: "LeagueClient:pid:porta:senha:https"
func parseLockfile(b []byte) (port, pass string, err error) {
	p := strings.Split(strings.TrimSpace(string(b)), ":")
	if len(p) < 5 {
		return "", "", errors.New("arquivo de conexão do LoL em formato inesperado")
	}
	if _, e := strconv.Atoi(p[2]); e != nil {
		return "", "", errors.New("porta inválida no arquivo de conexão do LoL")
	}
	return p[2], p[3], nil
}

// lugares onde o LoL costuma estar instalado
func lockfileCandidates() []string {
	var out []string
	add := func(dir string) {
		if dir != "" {
			out = append(out, filepath.Join(dir, "lockfile"))
		}
	}
	// o Riot Client anota onde instalou cada jogo
	pd := os.Getenv("ProgramData")
	if pd == "" {
		pd = `C:\ProgramData`
	}
	if b, err := os.ReadFile(filepath.Join(pd, "Riot Games", "RiotClientInstalls.json")); err == nil {
		var m map[string]any
		if json.Unmarshal(b, &m) == nil {
			if ac, ok := m["associated_client"].(map[string]any); ok {
				for dir := range ac {
					if strings.Contains(strings.ToLower(dir), "league") {
						add(filepath.FromSlash(dir))
					}
				}
			}
		}
	}
	for _, d := range []string{"C", "D", "E", "F", "G"} {
		add(d + `:\Riot Games\League of Legends`)
		add(d + `:\Program Files\Riot Games\League of Legends`)
		add(d + `:\Games\Riot Games\League of Legends`)
	}
	return out
}

// acha o cliente do LoL que está de fato aberto. Primeiro o lockfile da pasta de onde o LoL aberto
// está rodando (o atual), depois os lugares comuns. Cada candidato é conferido com um pedido de
// verdade (com a senha): um lockfile velho pode apontar pra uma porta que hoje é de outro programa,
// e aí o importar ficava esperando até estourar o tempo.
func lcuConnect() (*lcuConn, error) {
	type cand struct{ port, pass, from string }
	var cands []cand
	var lerr error
	seen := map[string]bool{}
	add := func(f string) {
		b, err := os.ReadFile(f)
		if err != nil {
			return
		}
		port, pass, err := parseLockfile(b)
		if err != nil {
			lerr = err
			return
		}
		if !seen[port+pass] {
			seen[port+pass] = true
			cands = append(cands, cand{port, pass, f})
		}
	}
	for _, d := range lcuProcessDirs() {
		add(filepath.Join(d, "lockfile"))
	}
	for _, f := range lockfileCandidates() {
		add(f)
	}
	anyAlive := false
	for _, c := range cands {
		cn, err := net.DialTimeout("tcp", "127.0.0.1:"+c.port, 700*time.Millisecond)
		if err != nil {
			lcuLog("  porta %s fechada (%s)", c.port, c.from)
			continue
		}
		cn.Close()
		anyAlive = true
		lc := newLcuConn("https://127.0.0.1:"+c.port, c.pass)
		lc.cl.Timeout = 4 * time.Second
		t0 := time.Now()
		st, err := lc.do1("GET", "/lol-gameflow/v1/gameflow-phase", nil, nil)
		if err == nil || (st >= 400 && st < 500 && st != 401 && st != 403) {
			lcuLog("  conectado na porta %s em %dms (%s)", c.port, time.Since(t0).Milliseconds(), c.from)
			lc.cl.Timeout = 10 * time.Second
			return lc, nil
		}
		lcuLog("  porta %s não é o LoL atual ou não respondeu: %d %v (%s)", c.port, st, err, c.from)
	}
	if anyAlive {
		return nil, errors.New("o LoL está aberto mas não respondeu agora (ele pode estar carregando ou reiniciando). Espere alguns segundos e tente de novo")
	}
	if len(cands) > 0 {
		return nil, errors.New("o LoL parece estar reiniciando (a porta dele não respondeu). Espere alguns segundos e tente de novo")
	}
	if lerr != nil {
		return nil, lerr
	}
	return nil, errors.New("não achei o LoL aberto neste PC. Abra o cliente do League of Legends e tente de novo")
}

// "--app-port=55549" e "--remoting-auth-token=xxxx" da linha de comando do LeagueClientUx.exe
func parseUxArgs(cmd string) (port, pass string) {
	grab := func(key string) string {
		i := strings.Index(cmd, key)
		if i < 0 {
			return ""
		}
		v := cmd[i+len(key):]
		v = strings.TrimLeft(v, `"`)
		if j := strings.IndexAny(v, "\" \t"); j >= 0 {
			v = v[:j]
		}
		return v
	}
	port, pass = grab("--app-port="), grab("--remoting-auth-token=")
	if _, err := strconv.Atoi(port); err != nil {
		return "", ""
	}
	return port, pass
}

func newLcuConn(base, pass string) *lcuConn {
	// o cliente usa um certificado próprio da Riot; a conexão é só com o próprio PC (127.0.0.1)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, Proxy: nil}
	return &lcuConn{base: base, pass: pass, cl: &http.Client{Timeout: 10 * time.Second, Transport: tr}}
}

// do faz a chamada e tenta de novo quando o cliente está ocupado (logo depois de uma partida o
// LoL às vezes demora ou responde 5xx por alguns segundos).
func (c *lcuConn) do(method, path string, body any, out any) (int, error) {
	var st int
	var err error
	for i := 0; i < 3; i++ {
		st, err = c.do1(method, path, body, out)
		if err == nil || !transient(st, err) {
			return st, err
		}
		time.Sleep(time.Duration(600*(i+1)) * time.Millisecond)
	}
	return st, err
}

// erro passageiro: sem resposta, demora ou 5xx (o cliente ainda está carregando)
func transient(st int, err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "demorou demais") { // já esperou o tempo todo: não repete
		return false
	}
	return st == 0 || st >= 500
}

// espera o cliente estar pronto pra receber pedidos (até ~6 s)
func (c *lcuConn) ready() error {
	var err error
	for i := 0; i < 4; i++ {
		var st int
		st, err = c.do1("GET", "/lol-summoner/v1/current-summoner", nil, nil)
		if err == nil || (st >= 400 && st < 500) {
			return nil
		}
		if !c.deadline.IsZero() && time.Until(c.deadline) < 2*time.Second {
			break
		}
		time.Sleep(1500 * time.Millisecond)
	}
	return fmt.Errorf("o cliente do LoL não está respondendo agora (se acabou de sair de uma partida, espere alguns segundos e tente de novo). Detalhe: %v", err)
}

func (c *lcuConn) do1(method, path string, body any, out any) (int, error) {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	ctx := context.Background()
	if !c.deadline.IsZero() {
		left := time.Until(c.deadline)
		if left <= 0 {
			return 0, errors.New("o LoL demorou demais pra responder (tente de novo em alguns segundos)")
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, left)
		defer cancel()
	}
	req, _ := http.NewRequestWithContext(ctx, method, c.base+path, rd)
	req.SetBasicAuth("riot", c.pass)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	t0 := time.Now()
	resp, err := c.cl.Do(req)
	if c.log {
		defer func(st *http.Response) {
			code := 0
			if st != nil {
				code = st.StatusCode
			}
			if err != nil {
				lcuLog("  %s %s -> erro em %dms: %v", method, path, time.Since(t0).Milliseconds(), err)
			} else {
				lcuLog("  %s %s -> %d em %dms", method, path, code, time.Since(t0).Milliseconds())
			}
		}(resp)
	}
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "deadline") || strings.Contains(err.Error(), "context") {
			return 0, errors.New("o LoL demorou demais pra responder (tente de novo em alguns segundos)")
		}
		return 0, fmt.Errorf("o LoL não respondeu (ele está aberto e logado?) [%v]", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if out != nil && resp.StatusCode < 300 && len(rb) > 0 {
		if err := json.Unmarshal(rb, out); err != nil {
			return resp.StatusCode, fmt.Errorf("resposta estranha do LoL em %s", path)
		}
	}
	if resp.StatusCode >= 300 {
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(rb, &e)
		return resp.StatusCode, fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, e.Message)
	}
	return resp.StatusCode, nil
}

// o que o app manda (montado no JS a partir da build)
type lcuBuild struct {
	Page *struct {
		Primary int   `json:"primaryStyleId"`
		Sub     int   `json:"subStyleId"`
		Perks   []int `json:"selectedPerkIds"`
	} `json:"page"`
	Spells   []int  `json:"spells"`
	Flash    string `json:"flash"` // keep | d | f
	ChampKey int    `json:"champKey"`
	Champ    string `json:"champ"`
	Title    string `json:"title"`
	Blocks   []struct {
		Type  string   `json:"type"`
		Items []string `json:"items"`
	} `json:"blocks"`
}

type lcuResult struct {
	Runes  string `json:"runes"`  // ok | full:<nome da página atual> | erro
	Spells string `json:"spells"` // ok | fora | erro | -
	Items  string `json:"items"`  // ok | erro | -
}

type perkPage struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	IsDeletable bool   `json:"isDeletable"`
	IsEditable  bool   `json:"isEditable"`
	Current     bool   `json:"current"`
}

func lcuImport(payload string, replace bool) (res lcuResult, err error) {
	t0 := time.Now()
	lcuLog("importar (trocar página: %v)", replace)
	defer func() {
		if err != nil {
			lcuLog("  FALHOU em %dms: %v", time.Since(t0).Milliseconds(), err)
		} else {
			lcuLog("  fim em %dms: runas=%s feitiços=%s itens=%s", time.Since(t0).Milliseconds(), res.Runes, res.Spells, res.Items)
		}
	}()
	c, err := lcuConnect()
	if err != nil {
		return lcuResult{}, err
	}
	// prazo total: o botão nunca fica preso esperando o LoL (o app espera até 35 s)
	c.deadline = time.Now().Add(26 * time.Second)
	c.cl.Timeout = 15 * time.Second
	c.log = true
	defer c.cl.CloseIdleConnections()
	return lcuImportWith(c, payload, replace)
}

// Runas, feitiços e itens são partes separadas do cliente: vão ao mesmo tempo. Antes iam uma
// depois da outra e, com o servidor da Riot lento, a soma passava do prazo (botão "travado").
func lcuImportWith(c *lcuConn, payload string, replace bool) (lcuResult, error) {
	var b lcuBuild
	if err := json.Unmarshal([]byte(payload), &b); err != nil {
		return lcuResult{}, errors.New("build inválida")
	}
	if err := c.ready(); err != nil {
		return lcuResult{}, err
	}
	res := lcuResult{Runes: "-", Spells: "-", Items: "-"}
	var wg sync.WaitGroup
	if b.Page != nil && len(b.Page.Perks) == 9 && b.Page.Primary > 0 && b.Page.Sub > 0 {
		wg.Add(1)
		go func() { defer wg.Done(); res.Runes = c.importRunes(b, replace) }()
	}
	if len(b.Spells) == 2 && b.Spells[0] > 0 && b.Spells[1] > 0 {
		wg.Add(1)
		go func() { defer wg.Done(); res.Spells = c.importSpells(b.Spells, b.Flash) }()
	}
	if b.ChampKey > 0 && len(b.Blocks) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.importItems(b); err != nil {
				res.Items = "erro: " + err.Error()
			} else {
				res.Items = "ok"
			}
		}()
	}
	wg.Wait()
	return res, nil
}

func (c *lcuConn) importRunes(b lcuBuild, replace bool) string {
	var pages []perkPage
	if _, err := c.do("GET", "/lol-perks/v1/pages", nil, &pages); err != nil {
		return "erro: " + err.Error()
	}
	for _, p := range pages {
		if p.Name == lcuPageName && p.IsDeletable {
			c.do("DELETE", "/lol-perks/v1/pages/"+strconv.Itoa(p.ID), nil, nil)
		}
	}
	body := map[string]any{"name": lcuPageName, "primaryStyleId": b.Page.Primary, "subStyleId": b.Page.Sub,
		"selectedPerkIds": b.Page.Perks, "current": true}
	_, err := c.do("POST", "/lol-perks/v1/pages", body, nil)
	if err == nil {
		return "ok"
	}
	// sem espaço pra página nova: só troca a página atual se a pessoa confirmou
	var cur perkPage
	c.do("GET", "/lol-perks/v1/currentpage", nil, &cur)
	if !replace {
		if cur.IsDeletable {
			return "full:" + cur.Name
		}
		return "erro: " + err.Error()
	}
	if !cur.IsDeletable {
		return "erro: a página atual não pode ser trocada"
	}
	if _, err := c.do("DELETE", "/lol-perks/v1/pages/"+strconv.Itoa(cur.ID), nil, nil); err != nil {
		return "erro: " + err.Error()
	}
	if _, err := c.do("POST", "/lol-perks/v1/pages", body, nil); err != nil {
		return "erro: " + err.Error()
	}
	return "ok"
}

func (c *lcuConn) importSpells(sp []int, flashKey string) string {
	var sess struct {
		LocalPlayerCellID int `json:"localPlayerCellId"`
		MyTeam            []struct {
			CellID int `json:"cellId"`
			Spell1 int `json:"spell1Id"`
			Spell2 int `json:"spell2Id"`
		} `json:"myTeam"`
	}
	code, err := c.do("GET", "/lol-champ-select/v1/session", nil, &sess)
	if code == 404 {
		return "fora"
	}
	if err != nil {
		return "erro: " + err.Error()
	}
	s1, s2 := sp[0], sp[1]
	const flash = 4
	if s1 == flash || s2 == flash {
		other := s1
		if s1 == flash {
			other = s2
		}
		switch flashKey {
		case "d": // spell1 = tecla D
			s1, s2 = flash, other
		case "f": // spell2 = tecla F
			s1, s2 = other, flash
		}
	}
	// "manter": deixa o Flash na tecla em que a pessoa já está usando
	for _, m := range sess.MyTeam {
		if flashKey == "d" || flashKey == "f" {
			break
		}
		if m.CellID != sess.LocalPlayerCellID {
			continue
		}
		if (s1 == flash || s2 == flash) && (m.Spell1 == flash || m.Spell2 == flash) {
			other := s1
			if s1 == flash {
				other = s2
			}
			if m.Spell1 == flash {
				s1, s2 = flash, other
			} else {
				s1, s2 = other, flash
			}
		}
	}
	if _, err := c.do("PATCH", "/lol-champ-select/v1/session/my-selection", map[string]int{"spell1Id": s1, "spell2Id": s2}, nil); err != nil {
		return "erro: " + err.Error()
	}
	return "ok"
}

func (c *lcuConn) importItems(b lcuBuild) error {
	var me struct {
		SummonerID int64 `json:"summonerId"`
	}
	if _, err := c.do("GET", "/lol-summoner/v1/current-summoner", nil, &me); err != nil {
		return err
	}
	if me.SummonerID == 0 {
		return errors.New("não achei sua conta no cliente")
	}
	path := "/lol-item-sets/v1/item-sets/" + strconv.FormatInt(me.SummonerID, 10) + "/sets"
	var all map[string]any
	if _, err := c.do("GET", path, nil, &all); err != nil {
		return err
	}
	if all == nil {
		all = map[string]any{}
	}
	uid := "ciencialol-" + strconv.Itoa(b.ChampKey)
	var keep []any
	minRank := 0.0
	if arr, ok := all["itemSets"].([]any); ok {
		for _, s := range arr {
			m, ok := s.(map[string]any)
			if ok && m["uid"] == uid {
				continue
			}
			if ok {
				if r, ok := m["sortrank"].(float64); ok && r < minRank {
					minRank = r
				}
			}
			keep = append(keep, s)
		}
	}
	var blocks []map[string]any
	for _, bl := range b.Blocks {
		var items []map[string]any
		for _, id := range bl.Items {
			if _, err := strconv.Atoi(id); err == nil {
				items = append(items, map[string]any{"id": id, "count": 1})
			}
		}
		if len(items) > 0 {
			blocks = append(blocks, map[string]any{"type": bl.Type, "items": items})
		}
	}
	// O arsenal da loja não respeitou a ordem da lista nem o sortrank alto: o "Blitz: ..." continuava
	// em cima. Pra garantir o primeiro lugar nos dois jeitos que o LoL pode ordenar:
	//  - nome começando com "(Ciência)": "(" vem antes de qualquer letra ou número na ordem alfabética;
	//  - sortrank menor que o de todos os outros conjuntos (ordem crescente).
	title := strings.TrimSpace(strings.TrimPrefix(b.Title, "Ciência LOL:"))
	if title == "" || title == "Ciência LOL" {
		title = b.Champ
	}
	title = "(Ciência) " + title
	if r := []rune(title); len(r) > 75 {
		title = string(r[:75])
	}
	keep = append([]any{map[string]any{
		"uid": uid, "title": title, "type": "custom", "map": "any", "mode": "any", "sortrank": int64(minRank) - 1,
		"startedFrom": "blank", "associatedChampions": []int{b.ChampKey}, "associatedMaps": []int{11, 12},
		"blocks": blocks, "preferredItemSlots": []any{},
	}}, keep...)
	all["itemSets"] = keep
	all["timestamp"] = time.Now().UnixMilli()
	_, err := c.do("PUT", path, all, nil)
	return err
}

// ---------------------------------------------------------------------------
// Estatísticas: lê (só leitura) as últimas partidas da conta logada no cliente.
// Nada é alterado no cliente; o resultado (vitória/derrota) continua sendo marcado à mão no app.

type lcuGameStats struct {
	GameID   int64  `json:"gid"`
	Creation int64  `json:"at"`  // início da partida (ms)
	Duration int    `json:"dur"` // segundos
	Queue    int    `json:"q"`
	Champ    int    `json:"champ"`
	Win      bool   `json:"win"`
	K        int    `json:"k"`
	D        int    `json:"d"`
	A        int    `json:"a"`
	Dmg      int    `json:"dmg"`
	Gold     int    `json:"gold"`
	CS       int    `json:"cs"`
	Vision   int    `json:"vis"`
	Level    int    `json:"lvl"`
	Items    []int  `json:"items"`
	Mode     string `json:"mode,omitempty"`
}

type lcuMHGame struct {
	GameID       int64  `json:"gameId"`
	GameCreation int64  `json:"gameCreation"`
	GameDuration int    `json:"gameDuration"`
	QueueID      int    `json:"queueId"`
	GameMode     string `json:"gameMode"`
	Participants []struct {
		ParticipantID int `json:"participantId"`
		ChampionID    int `json:"championId"`
		Stats         struct {
			Win                         bool `json:"win"`
			Kills                       int  `json:"kills"`
			Deaths                      int  `json:"deaths"`
			Assists                     int  `json:"assists"`
			TotalDamageDealtToChampions int  `json:"totalDamageDealtToChampions"`
			GoldEarned                  int  `json:"goldEarned"`
			TotalMinionsKilled          int  `json:"totalMinionsKilled"`
			NeutralMinionsKilled        int  `json:"neutralMinionsKilled"`
			VisionScore                 int  `json:"visionScore"`
			ChampLevel                  int  `json:"champLevel"`
			Item0                       int  `json:"item0"`
			Item1                       int  `json:"item1"`
			Item2                       int  `json:"item2"`
			Item3                       int  `json:"item3"`
			Item4                       int  `json:"item4"`
			Item5                       int  `json:"item5"`
			Item6                       int  `json:"item6"`
		} `json:"stats"`
	} `json:"participants"`
	ParticipantIdentities []struct {
		ParticipantID int `json:"participantId"`
		Player        struct {
			Puuid string `json:"puuid"`
		} `json:"player"`
	} `json:"participantIdentities"`
}

func lcuRecent() ([]lcuGameStats, error) {
	c, err := lcuConnect()
	if err != nil {
		return nil, err
	}
	// o histórico de partidas o cliente busca nos servidores da Riot: pode demorar bem mais que o resto
	c.cl.Timeout = 45 * time.Second
	return lcuRecentWith(c)
}

func lcuRecentWith(c *lcuConn) ([]lcuGameStats, error) {
	var me struct {
		Puuid string `json:"puuid"`
	}
	if _, err := c.do("GET", "/lol-summoner/v1/current-summoner", nil, &me); err != nil {
		return nil, err
	}
	// 1) resumo da tela de fim de partida (fica no próprio cliente, responde na hora)
	eog, eogErr := c.eogStats()
	// 2) histórico de partidas (o cliente pergunta pros servidores da Riot)
	var mh struct {
		Games struct {
			Games []lcuMHGame `json:"games"`
		} `json:"games"`
	}
	_, err := c.do("GET", "/lol-match-history/v1/products/lol/current-summoner/matches?begIndex=0&endIndex=6", nil, &mh)
	if err != nil && me.Puuid != "" {
		_, err = c.do("GET", "/lol-match-history/v1/products/lol/"+me.Puuid+"/matches?begIndex=0&endIndex=6", nil, &mh)
	}
	if err != nil {
		if eog != nil {
			return []lcuGameStats{*eog}, nil
		}
		if eogErr != nil {
			return nil, err
		}
		return nil, err
	}
	var out []lcuGameStats
	if eog != nil {
		dup := false
		for _, g := range mh.Games.Games {
			if g.GameID == eog.GameID {
				dup = true
			}
		}
		if !dup {
			out = append(out, *eog)
		}
	}
	for _, g := range mh.Games.Games {
		// acha o participante da conta logada (normalmente é o único que vem nessa lista)
		pid := 0
		for _, id := range g.ParticipantIdentities {
			if me.Puuid != "" && id.Player.Puuid == me.Puuid {
				pid = id.ParticipantID
			}
		}
		for _, p := range g.Participants {
			if pid != 0 && p.ParticipantID != pid {
				continue
			}
			s := p.Stats
			out = append(out, lcuGameStats{GameID: g.GameID, Creation: g.GameCreation, Duration: g.GameDuration, Queue: g.QueueID, Mode: g.GameMode,
				Champ: p.ChampionID, Win: s.Win, K: s.Kills, D: s.Deaths, A: s.Assists, Dmg: s.TotalDamageDealtToChampions, Gold: s.GoldEarned,
				CS: s.TotalMinionsKilled + s.NeutralMinionsKilled, Vision: s.VisionScore, Level: s.ChampLevel,
				Items: []int{s.Item0, s.Item1, s.Item2, s.Item3, s.Item4, s.Item5, s.Item6}})
			break
		}
	}
	return out, nil
}

// resumo da tela de fim de partida (só existe depois que uma partida termina, até o cliente limpar)
func (c *lcuConn) eogStats() (*lcuGameStats, error) {
	var e struct {
		GameID      int64  `json:"gameId"`
		GameLength  int    `json:"gameLength"`
		GameMode    string `json:"gameMode"`
		QueueID     int    `json:"queueId"`
		LocalPlayer struct {
			ChampionID int            `json:"championId"`
			Items      []int          `json:"items"`
			Stats      map[string]int `json:"stats"`
		} `json:"localPlayer"`
	}
	code, err := c.do("GET", "/lol-end-of-game/v1/eog-stats-block", nil, &e)
	if err != nil || code != 200 || e.GameID == 0 || e.LocalPlayer.ChampionID == 0 {
		return nil, err
	}
	st := e.LocalPlayer.Stats
	g := &lcuGameStats{GameID: e.GameID, Duration: e.GameLength, Queue: e.QueueID, Mode: e.GameMode, Champ: e.LocalPlayer.ChampionID,
		Creation: time.Now().UnixMilli() - int64(e.GameLength)*1000 - 60000,
		Win:      st["WIN"] == 1, K: st["CHAMPIONS_KILLED"], D: st["NUM_DEATHS"], A: st["ASSISTS"],
		Dmg: st["TOTAL_DAMAGE_DEALT_TO_CHAMPIONS"], Gold: st["GOLD_EARNED"], CS: st["MINIONS_KILLED"] + st["NEUTRAL_MINIONS_KILLED"],
		Vision: st["VISION_SCORE"], Level: st["LEVEL"], Items: e.LocalPlayer.Items}
	return g, nil
}

// ---- fase do cliente (lobby, partida, fim de jogo...) pra o app saber quando a partida acabou ----
var (
	phaseMu   sync.Mutex
	phaseConn *lcuConn
	phaseFull time.Time // última vez que procurou o LoL pelo processo (mais pesado)
)

func lcuPhase() (string, error) {
	phaseMu.Lock()
	defer phaseMu.Unlock()
	try := func(c *lcuConn) (string, bool) {
		var ph string
		c.cl.Timeout = 3 * time.Second
		if _, err := c.do1("GET", "/lol-gameflow/v1/gameflow-phase", nil, &ph); err != nil {
			return "", false
		}
		return ph, true
	}
	if phaseConn != nil {
		if ph, ok := try(phaseConn); ok {
			return ph, nil
		}
		phaseConn = nil
	}
	// procura de novo: pelo processo no máximo 1 vez por minuto (o resto do tempo, só pelo lockfile)
	var c *lcuConn
	var err error
	if time.Since(phaseFull) > time.Minute {
		phaseFull = time.Now()
		c, err = lcuConnect()
	} else {
		c, err = lcuConnectLockfile()
	}
	if err != nil {
		return "", nil // LoL fechado: sem fase
	}
	if ph, ok := try(c); ok {
		phaseConn = c
		return ph, nil
	}
	return "", nil
}

// só pelos lockfiles (rápido, sem consultar processos)
func lcuConnectLockfile() (*lcuConn, error) {
	for _, f := range lockfileCandidates() {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		port, pass, err := parseLockfile(b)
		if err != nil {
			continue
		}
		cn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 500*time.Millisecond)
		if err != nil {
			continue
		}
		cn.Close()
		return newLcuConn("https://127.0.0.1:"+port, pass), nil
	}
	return nil, errors.New("sem LoL")
}

// ---- registro do importar (lcu.log na pasta de dados): mostra onde cada tentativa parou ----
var (
	lcuLogPath string
	lcuLogMu   sync.Mutex
)

func lcuLogInit(dir string) { lcuLogPath = filepath.Join(dir, "lcu.log") }

func lcuLog(format string, a ...any) {
	if lcuLogPath == "" {
		return
	}
	lcuLogMu.Lock()
	defer lcuLogMu.Unlock()
	if fi, err := os.Stat(lcuLogPath); err == nil && fi.Size() > 512<<10 {
		_ = os.Rename(lcuLogPath, lcuLogPath+".old")
	}
	f, err := os.OpenFile(lcuLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s  %s\n", time.Now().Format("2006-01-02 15:04:05.000"), fmt.Sprintf(format, a...))
}
