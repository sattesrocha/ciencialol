package main

// Leitura do op.gg pelo servidor público oficial deles para programas (OP.GG MCP,
// https://mcp-api.op.gg/mcp). O app descobre as ferramentas e os parâmetros na hora
// (tools/list), então pequenas mudanças do lado deles não quebram tudo. Tudo que vem
// de lá fica registrado em opgg.log (pasta de dados), pra ajustar se o formato mudar.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var opggURL = "https://mcp-api.op.gg/mcp"
var opggLogPath string

type mcpTool struct {
	Name        string `json:"name"`
	InputSchema struct {
		Properties map[string]struct {
			Type        any    `json:"type"`
			Enum        []any  `json:"enum"`
			Description string `json:"description"`
			Items       *struct {
				Enum []any `json:"enum"`
			} `json:"items"`
		} `json:"properties"`
		Required []string `json:"required"`
	} `json:"inputSchema"`
}

type mcpClient struct {
	url string
	sid string
	id  int
	hc  *http.Client
}

var (
	opggMu    sync.Mutex
	opggTools []mcpTool
	opggAt    time.Time
)

func opggLog(what string, b []byte) {
	if opggLogPath == "" {
		return
	}
	if st, err := os.Stat(opggLogPath); err == nil && st.Size() > 2<<20 {
		_ = os.Remove(opggLogPath)
	}
	f, err := os.OpenFile(opggLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	if len(b) > 200000 {
		b = b[:200000]
	}
	fmt.Fprintf(f, "\n==== %s %s\n%s\n", time.Now().Format("2006-01-02 15:04:05"), what, b)
}

func (c *mcpClient) rpc(method string, params any, notify bool) (json.RawMessage, error) {
	msg := map[string]any{"jsonrpc": "2.0", "method": method}
	if params != nil {
		msg["params"] = params
	}
	myID := 0
	if !notify {
		c.id++
		myID = c.id
		msg["id"] = myID
	}
	body, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", c.url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-03-26")
	req.Header.Set("User-Agent", "CienciaLOL/"+AppVersion)
	if c.sid != "" {
		req.Header.Set("Mcp-Session-Id", c.sid)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sem conexão com o op.gg (%v)", err)
	}
	defer resp.Body.Close()
	if s := resp.Header.Get("Mcp-Session-Id"); s != "" {
		c.sid = s
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if notify {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		opggLog(method+" HTTP "+strconv.Itoa(resp.StatusCode), raw)
		return nil, fmt.Errorf("o op.gg respondeu com erro %d", resp.StatusCode)
	}
	// a resposta pode vir como JSON direto ou como "event stream" (linhas data: {...})
	var cands [][]byte
	if strings.Contains(resp.Header.Get("Content-Type"), "event-stream") || bytes.HasPrefix(bytes.TrimSpace(raw), []byte("event:")) || bytes.HasPrefix(bytes.TrimSpace(raw), []byte("data:")) {
		sc := bufio.NewScanner(bytes.NewReader(raw))
		sc.Buffer(make([]byte, 1<<20), 16<<20)
		var cur bytes.Buffer
		flush := func() {
			if cur.Len() > 0 {
				cands = append(cands, append([]byte(nil), cur.Bytes()...))
				cur.Reset()
			}
		}
		for sc.Scan() {
			ln := sc.Text()
			if strings.HasPrefix(ln, "data:") {
				cur.WriteString(strings.TrimSpace(strings.TrimPrefix(ln, "data:")))
			} else if strings.TrimSpace(ln) == "" {
				flush()
			}
		}
		flush()
	} else {
		cands = [][]byte{raw}
	}
	for _, cnd := range cands {
		var r struct {
			ID     any             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(cnd, &r) != nil {
			continue
		}
		if fmt.Sprint(r.ID) != strconv.Itoa(myID) {
			continue
		}
		if r.Error != nil {
			opggLog(method+" erro", cnd)
			return nil, errors.New("o op.gg recusou: " + r.Error.Message)
		}
		return r.Result, nil
	}
	opggLog(method+" sem resposta reconhecida", raw)
	return nil, errors.New("resposta do op.gg em formato desconhecido (detalhe no opgg.log)")
}

func opggSession() (*mcpClient, []mcpTool, error) {
	c := &mcpClient{url: opggURL, hc: &http.Client{Timeout: 25 * time.Second}}
	if _, err := c.rpc("initialize", map[string]any{"protocolVersion": "2025-03-26", "capabilities": map[string]any{},
		"clientInfo": map[string]any{"name": "CienciaLOL", "version": AppVersion}}, false); err != nil {
		return nil, nil, err
	}
	_, _ = c.rpc("notifications/initialized", nil, true)
	opggMu.Lock()
	tools := opggTools
	fresh := time.Since(opggAt) < 6*time.Hour
	opggMu.Unlock()
	if !fresh || len(tools) == 0 {
		raw, err := c.rpc("tools/list", map[string]any{}, false)
		if err != nil {
			return nil, nil, err
		}
		var tl struct {
			Tools []mcpTool `json:"tools"`
		}
		if err := json.Unmarshal(raw, &tl); err != nil || len(tl.Tools) == 0 {
			opggLog("tools/list", raw)
			return nil, nil, errors.New("o op.gg não listou as ferramentas")
		}
		tools = tl.Tools
		opggMu.Lock()
		opggTools, opggAt = tools, time.Now()
		opggMu.Unlock()
	}
	return c, tools, nil
}

// escolhe a ferramenta cujo nome tem todas as palavras de um dos grupos (em ordem de preferência)
func pickTool(tools []mcpTool, groups ...[]string) *mcpTool {
	for _, g := range groups {
		for i := range tools {
			n := strings.ToLower(strings.ReplaceAll(tools[i].Name, "-", "_"))
			ok := true
			for _, w := range g {
				if !strings.Contains(n, w) {
					ok = false
					break
				}
			}
			if ok {
				return &tools[i]
			}
		}
	}
	return nil
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]`)

func normKey(s string) string { return nonAlnum.ReplaceAllString(strings.ToLower(s), "") }

// enumPick devolve o valor da lista que bate com algum dos desejos (ignorando maiúsculas e símbolos)
func enumPick(enum []any, wants ...string) (any, bool) {
	for _, w := range wants {
		wk := normKey(w)
		for _, e := range enum {
			if normKey(fmt.Sprint(e)) == wk {
				return e, true
			}
		}
	}
	for _, w := range wants {
		wk := normKey(w)
		for _, e := range enum {
			if ek := normKey(fmt.Sprint(e)); wk != "" && (strings.HasPrefix(ek, wk) || strings.HasPrefix(wk, ek)) {
				return e, true
			}
		}
	}
	return nil, false
}

// buildArgs monta os parâmetros a partir do que a ferramenta pede
func buildArgs(t *mcpTool, vals map[string][]string) map[string]any {
	args := map[string]any{}
	req := map[string]bool{}
	for _, r := range t.InputSchema.Required {
		req[r] = true
	}
	for name, p := range t.InputSchema.Properties {
		k := normKey(name)
		var wants []string
		switch {
		case k == "gamename" || k == "summonername" || k == "name" && vals["gamename"] != nil:
			wants = vals["gamename"]
		case k == "tagline" || k == "tag":
			wants = vals["tagline"]
		case k == "region" || k == "platform":
			wants = []string{"br", "br1"}
		case k == "lang" || k == "language" || k == "locale" || k == "hl":
			wants = []string{"pt_BR", "pt-BR", "ptbr", "en_US", "en"}
		case k == "champion" || k == "championname" || k == "champ":
			wants = vals["champion"]
		case k == "position" || k == "lane" || k == "role":
			wants = vals["position"]
		case k == "gamemode" || k == "mode" || k == "queue" || k == "queuetype":
			wants = []string{"ranked", "RANKED", "SOLORANKED", "rank"}
		case k == "limit" || k == "count" || k == "size":
			args[name] = 20
			continue
		case k == "desiredoutputfields":
			continue // sem isso a ferramenta devolve os campos padrão
		default:
			if req[name] && len(p.Enum) > 0 {
				args[name] = p.Enum[0]
			}
			continue
		}
		if len(wants) == 0 {
			continue
		}
		if len(p.Enum) > 0 {
			if v, ok := enumPick(p.Enum, wants...); ok {
				args[name] = v
			} else if req[name] {
				args[name] = p.Enum[0]
			}
			continue
		}
		args[name] = wants[0]
	}
	return args
}

func opggCall(c *mcpClient, t *mcpTool, args map[string]any) (any, error) {
	ab, _ := json.Marshal(args)
	opggLog("chamando "+t.Name, ab)
	raw, err := c.rpc("tools/call", map[string]any{"name": t.Name, "arguments": args}, false)
	if err != nil {
		return nil, err
	}
	opggLog("resposta "+t.Name, raw)
	var r struct {
		IsError    bool            `json:"isError"`
		Structured json.RawMessage `json:"structuredContent"`
		Content    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, errors.New("resposta do op.gg em formato desconhecido")
	}
	if r.IsError {
		msg := ""
		if len(r.Content) > 0 {
			msg = r.Content[0].Text
		}
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, errors.New("o op.gg não achou: " + msg)
	}
	var v any
	if len(r.Structured) > 0 && json.Unmarshal(r.Structured, &v) == nil {
		return v, nil
	}
	for _, ct := range r.Content {
		txt := strings.TrimSpace(ct.Text)
		if json.Unmarshal([]byte(txt), &v) == nil {
			return v, nil
		}
		// formato compacto do op.gg (classes + valores em ordem)
		if strings.HasPrefix(txt, "class ") {
			if pv, err := parseOpggText(txt); err == nil {
				return pv, nil
			} else {
				opggLog("não entendi o formato", []byte(err.Error()))
			}
		}
		// às vezes vem texto antes do JSON
		if i := strings.IndexAny(txt, "{["); i >= 0 && json.Unmarshal([]byte(txt[i:]), &v) == nil {
			return v, nil
		}
	}
	return nil, errors.New("o op.gg respondeu num formato que o app ainda não entende (detalhe no opgg.log)")
}

// ---------- leitura tolerante do JSON do op.gg ----------

func num(v any) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(x))
		return n, err == nil
	}
	return 0, false
}

func intsOf(v any) []int {
	var out []int
	switch x := v.(type) {
	case []any:
		for _, e := range x {
			if n, ok := num(e); ok && n > 0 {
				out = append(out, n)
			} else if m, ok := e.(map[string]any); ok {
				for _, k := range []string{"id", "item_id", "itemId", "spell_id", "rune_id"} {
					if n, ok := num(m[k]); ok && n > 0 {
						out = append(out, n)
						break
					}
				}
			}
		}
	}
	return out
}

func getAny(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	// sem ligar pra maiúsculas/sublinhado
	for mk, v := range m {
		nk := normKey(mk)
		for _, k := range keys {
			if nk == normKey(k) && v != nil {
				return v
			}
		}
	}
	return nil
}

func champOf(m map[string]any) (int, string) {
	if n, ok := num(getAny(m, "champion_id", "championId")); ok {
		return n, ""
	}
	switch c := getAny(m, "champion", "champion_name", "championName").(type) {
	case float64:
		return int(c), ""
	case string:
		return 0, c
	case map[string]any:
		id, _ := num(getAny(c, "id", "key", "champion_id"))
		nm, _ := getAny(c, "name", "key_name", "en_name").(string)
		return id, nm
	}
	return 0, ""
}

func runesOf(m map[string]any, b *extBuild) {
	r, _ := getAny(m, "rune", "runes", "rune_page", "perks").(map[string]any)
	if r == nil {
		r = m
	}
	if arr, ok := getAny(m, "runes", "rune_pages").([]any); ok && len(arr) > 0 {
		if x, ok := arr[0].(map[string]any); ok {
			r = x
		}
	}
	b.Primary, _ = num(getAny(r, "primary_page_id", "primaryStyleId", "primary_style_id", "perkPrimaryStyle"))
	b.Sub, _ = num(getAny(r, "secondary_page_id", "subStyleId", "sub_style_id", "secondary_style_id", "perkSubStyle"))
	prim := intsOf(getAny(r, "primary_rune_ids", "primary_runes"))
	if len(prim) == 0 {
		if n, ok := num(getAny(r, "primary_rune_id", "keystone_id")); ok && n > 0 {
			prim = []int{n}
		}
	}
	sec := intsOf(getAny(r, "secondary_rune_ids", "secondary_runes"))
	sh := intsOf(getAny(r, "stat_mod_ids", "stat_perks", "shards"))
	b.Perks = append(append(append([]int{}, prim...), sec...), sh...)
	if len(prim) == 4 && len(sec) == 2 && len(sh) != 3 {
		b.Perks = append(append([]int{}, prim...), sec...)
	}
}

func winOf(m map[string]any) *bool {
	var w bool
	switch v := getAny(m, "win", "is_win", "isWin").(type) {
	case bool:
		w = v
		return &w
	}
	res := getAny(m, "result")
	if st, ok := getAny(m, "stats").(map[string]any); ok && res == nil {
		res = getAny(st, "result", "win")
	}
	switch v := res.(type) {
	case string:
		w = strings.EqualFold(v, "WIN") || strings.EqualFold(v, "victory")
		return &w
	case bool:
		w = v
		return &w
	}
	return nil
}

// acha as partidas: objetos que têm campeão + itens (de preferência os "meus")
func opggFindGames(v any, gameName string) []extBuild {
	var mine, all []extBuild
	var walk func(v any, path string, ctx map[string]any)
	walk = func(v any, path string, ctx map[string]any) {
		switch x := v.(type) {
		case map[string]any:
			if getAny(x, "created_at", "game_created_at", "gameCreation") != nil {
				ctx = x
			}
			items := intsOf(getAny(x, "items", "item_ids", "itemIds", "final_items"))
			id, nm := champOf(x)
			if len(items) > 0 && (id > 0 || nm != "") {
				b := extBuild{Src: "opgg", ChampKey: id, Champ: nm, Items: items, Win: winOf(x)}
				b.Spells = intsOf(getAny(x, "spells", "spell_ids", "summoner_spells"))
				runesOf(x, &b)
				if st, ok := getAny(x, "stats").(map[string]any); ok {
					b.K, _ = num(getAny(st, "kill", "kills"))
					b.D, _ = num(getAny(st, "death", "deaths"))
					b.A, _ = num(getAny(st, "assist", "assists"))
				}
				if ctx != nil {
					switch t := getAny(ctx, "created_at", "game_created_at", "gameCreation").(type) {
					case string:
						if tm, err := time.Parse(time.RFC3339, t); err == nil {
							b.At = tm.UnixMilli()
						}
					case float64:
						b.At = int64(t)
						if b.At < 1e12 {
							b.At *= 1000
						}
					}
					if gm, ok := getAny(ctx, "game_type", "queue_type", "game_mode").(string); ok {
						b.Mode = gm
					}
				}
				isMine := strings.Contains(strings.ToLower(path), "my") // myData / my_data
				if s, ok := getAny(x, "summoner").(map[string]any); ok && gameName != "" {
					if gn, _ := getAny(s, "game_name", "gameName").(string); strings.EqualFold(gn, gameName) {
						isMine = true
					}
				}
				if isMine {
					mine = append(mine, b)
				} else {
					all = append(all, b)
				}
				return
			}
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				walk(x[k], path+"."+k, ctx)
			}
		case []any:
			for _, e := range x {
				walk(e, path+"[]", ctx)
			}
		}
	}
	walk(v, "", nil)
	if len(mine) > 0 {
		return mine
	}
	return all
}

// build recomendada: primeira página de runas, feitiços, iniciais, botas e núcleo
func opggFindRec(v any) (extBuild, bool) {
	b := extBuild{Src: "opggrec"}
	found := false
	firstIDs := func(x any) []int {
		switch t := x.(type) {
		case []any:
			if len(t) == 0 {
				return nil
			}
			if m, ok := t[0].(map[string]any); ok {
				if ids := intsOf(getAny(m, "ids", "item_ids", "items", "spell_ids")); len(ids) > 0 {
					if pr := getAny(m, "pick_rate", "pickRate", "play"); pr != nil {
						b.Rate = fmt.Sprint(pr)
					}
					return ids
				}
			}
			return intsOf(t)
		case map[string]any:
			return intsOf(getAny(t, "ids", "item_ids", "items"))
		}
		return nil
	}
	var walk func(v any)
	walk = func(v any) {
		m, ok := v.(map[string]any)
		if !ok {
			if a, ok := v.([]any); ok {
				for _, e := range a {
					walk(e)
				}
			}
			return
		}
		for k, val := range m {
			nk := normKey(k)
			switch {
			case b.Primary == 0 && (nk == "runes" || nk == "runepages" || nk == "rune" || nk == "runepage"):
				tmp := extBuild{}
				runesOf(map[string]any{k: val}, &tmp)
				if tmp.Primary > 0 && len(tmp.Perks) >= 6 {
					b.Primary, b.Sub, b.Perks = tmp.Primary, tmp.Sub, tmp.Perks
					found = true
				}
			case nk == "positions" && len(b.Alt) == 0:
				if a, ok := val.([]any); ok {
					for _, e := range a {
						if pm, ok := e.(map[string]any); ok {
							if n, ok := pm["name"].(string); ok && n != "" {
								b.Alt = append(b.Alt, strings.ToLower(n))
							}
						}
					}
				}
			case len(b.Skills) == 0 && (nk == "skills" || nk == "skillorder"):
				sm, _ := val.(map[string]any)
				if a, ok := val.([]any); ok && len(a) > 0 {
					sm, _ = a[0].(map[string]any)
				}
				if sm != nil {
					if ord, ok := sm["order"].([]any); ok {
						for _, o := range ord {
							if k, ok := o.(string); ok && len(k) == 1 && strings.Contains("QWER", k) {
								b.Skills = append(b.Skills, k)
							}
						}
					}
				}
			case len(b.Spells) == 0 && (nk == "summonerspells" || nk == "spells"):
				if ids := firstIDs(val); len(ids) == 2 {
					b.Spells = ids
				}
			case len(b.Start) == 0 && (nk == "starteritems" || nk == "startitems"):
				b.Start = firstIDs(val)
			case len(b.Boots) == 0 && nk == "boots":
				b.Boots = firstIDs(val)
			case len(b.Items) == 0 && (nk == "coreitems" || nk == "core"):
				b.Items = firstIDs(val)
				cm, _ := val.(map[string]any)
				if a, ok := val.([]any); ok && len(a) > 0 {
					cm, _ = a[0].(map[string]any)
				}
				if cm != nil {
					b.Play, _ = num(cm["play"])
				}
				if len(b.Items) > 0 {
					found = true
				}
			}
		}
		if b.Primary == 0 {
			if _, ok := m["primary_page_id"]; ok {
				runesOf(m, &b)
				found = found || b.Primary > 0
			}
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			walk(m[k])
		}
	}
	walk(v)
	return b, found
}

var riotIDRx = regexp.MustCompile(`^\s*([^#]{2,32})\s*#\s*([A-Za-z0-9]{2,6})\s*$`)

// opggMatches: build das últimas partidas do jogador (Riot ID "nome#tag") no op.gg
func opggMatches(riotID string) ([]extBuild, error) {
	m := riotIDRx.FindStringSubmatch(riotID)
	if m == nil {
		return nil, errors.New("Riot ID inválido (use nome#tag, ex.: Jogador Teste#br1)")
	}
	c, tools, err := opggSession()
	if err != nil {
		return nil, err
	}
	vals := map[string][]string{"gamename": {strings.TrimSpace(m[1])}, "tagline": {m[2]}}
	// pede pro op.gg atualizar o perfil antes (se a ferramenta existir); não para se falhar
	if t := pickTool(tools, []string{"summoner", "renew"}); t != nil {
		_, _ = opggCall(c, t, buildArgs(t, vals))
	}
	t := pickTool(tools, []string{"summoner", "matches"}, []string{"summoner", "game", "history"}, []string{"summoner", "history"}, []string{"match", "list"})
	if t == nil {
		return nil, errors.New("o op.gg não oferece mais o histórico de partidas")
	}
	v, err := opggCall(c, t, buildArgs(t, vals))
	if err != nil {
		return nil, err
	}
	g := opggFindGames(v, strings.TrimSpace(m[1]))
	if len(g) == 0 {
		return nil, errors.New("não achei partidas com itens nessa resposta do op.gg (detalhe no opgg.log)")
	}
	if len(g) > 20 {
		g = g[:20]
	}
	return g, nil
}

// opggChampBuild: build recomendada do op.gg pro campeão e a função
func opggChampBuild(champ, champID, pos string) (extBuild, error) {
	c, tools, err := opggSession()
	if err != nil {
		return extBuild{}, err
	}
	t := pickTool(tools, []string{"champion", "analysis"}, []string{"champion", "build"})
	if t == nil {
		return extBuild{}, errors.New("o op.gg não oferece mais a build de campeões")
	}
	posW := map[string][]string{"top": {"top"}, "jungle": {"jungle", "jg"}, "mid": {"mid", "middle"}, "adc": {"adc", "bottom", "bot"}, "sup": {"support", "sup", "utility"}}[pos]
	if posW == nil {
		posW = []string{"mid"}
	}
	vals := map[string][]string{"champion": {champID, champ, strings.ToUpper(champID)}, "position": posW}
	v, err := opggCall(c, t, buildArgs(t, vals))
	if err != nil {
		return extBuild{}, err
	}
	b, ok := opggFindRec(v)
	if !ok {
		return extBuild{}, errors.New("não achei a build nessa resposta do op.gg (detalhe no opgg.log)")
	}
	b.Champ, b.Pos = champ, pos
	return b, nil
}

func opggInit(dir string) { opggLogPath = filepath.Join(dir, "opgg.log") }

// ---------- formato compacto do op.gg ----------
// O servidor do op.gg responde num texto enxuto: primeiro as "classes" com os nomes dos campos
// (class Runes: id,primary_page_id,...) e depois os valores em ordem, como Runes(8010,8000,...).
// parseOpggText transforma isso em mapas normais (campo -> valor).

func parseOpggText(txt string) (any, error) {
	classes := map[string][]string{}
	body := []string{}
	for _, ln := range strings.Split(txt, "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "class ") {
			if i := strings.Index(t, ":"); i > 6 {
				name := strings.TrimSpace(t[6:i])
				var fs []string
				for _, f := range strings.Split(t[i+1:], ",") {
					fs = append(fs, strings.TrimSpace(f))
				}
				classes[name] = fs
			}
			continue
		}
		body = append(body, ln)
	}
	p := &opggParser{s: strings.Join(body, "\n"), classes: classes}
	p.ws()
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	return v, nil
}

type opggParser struct {
	s       string
	i       int
	classes map[string][]string
}

func (p *opggParser) ws() {
	for p.i < len(p.s) && strings.ContainsRune(" \t\r\n", rune(p.s[p.i])) {
		p.i++
	}
}

func (p *opggParser) errf(m string) error { return fmt.Errorf("op.gg: %s (posição %d)", m, p.i) }

func (p *opggParser) value() (any, error) {
	p.ws()
	if p.i >= len(p.s) {
		return nil, p.errf("fim inesperado")
	}
	c := p.s[p.i]
	switch {
	case c == '"':
		j := p.i + 1
		for j < len(p.s) && p.s[j] != '"' {
			if p.s[j] == '\\' {
				j++
			}
			j++
		}
		if j >= len(p.s) {
			return nil, p.errf("texto sem fim")
		}
		var out string
		if err := json.Unmarshal([]byte(p.s[p.i:j+1]), &out); err != nil {
			out = p.s[p.i+1 : j]
		}
		p.i = j + 1
		return out, nil
	case c == '[':
		p.i++
		arr := []any{}
		for {
			p.ws()
			if p.i < len(p.s) && p.s[p.i] == ']' {
				p.i++
				return arr, nil
			}
			v, err := p.value()
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
			p.ws()
			if p.i < len(p.s) && p.s[p.i] == ',' {
				p.i++
			}
		}
	case c == '{':
		// JSON comum no meio (por garantia)
		dec := json.NewDecoder(strings.NewReader(p.s[p.i:]))
		var v any
		if err := dec.Decode(&v); err != nil {
			return nil, p.errf("objeto inválido")
		}
		p.i += int(dec.InputOffset())
		return v, nil
	case c == '-' || (c >= '0' && c <= '9'):
		j := p.i + 1
		for j < len(p.s) && strings.ContainsRune("0123456789.eE+-", rune(p.s[j])) {
			j++
		}
		f, err := strconv.ParseFloat(p.s[p.i:j], 64)
		if err != nil {
			return nil, p.errf("número inválido")
		}
		p.i = j
		return f, nil
	}
	// identificador: true/false/null ou Classe(...)
	j := p.i
	for j < len(p.s) && (p.s[j] == '_' || p.s[j] >= 'a' && p.s[j] <= 'z' || p.s[j] >= 'A' && p.s[j] <= 'Z' || p.s[j] >= '0' && p.s[j] <= '9') {
		j++
	}
	if j == p.i {
		return nil, p.errf("caractere inesperado " + strconv.Quote(string(c)))
	}
	id := p.s[p.i:j]
	p.i = j
	switch id {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null", "None":
		return nil, nil
	}
	p.ws()
	if p.i >= len(p.s) || p.s[p.i] != '(' {
		return id, nil
	}
	p.i++
	var args []any
	for {
		p.ws()
		if p.i < len(p.s) && p.s[p.i] == ')' {
			p.i++
			break
		}
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		args = append(args, v)
		p.ws()
		if p.i < len(p.s) && p.s[p.i] == ',' {
			p.i++
		}
	}
	fs := p.classes[id]
	m := map[string]any{"_class": id}
	for k, a := range args {
		if k < len(fs) && fs[k] != "" {
			m[fs[k]] = a
		} else {
			m["_"+strconv.Itoa(k)] = a
		}
	}
	return m, nil
}
