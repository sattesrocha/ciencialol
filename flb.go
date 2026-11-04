package main

// FunLeagueBuilds (funleaguebuilds.de)
// A busca na API do site fica DESLIGADA até o FunLeagueBuilds autorizar.
// Quando autorizarem: FLBLive = true, gerar nova versão.
// Regras combinadas no pedido: no máximo 1 atualização por dia (2 requisições),
// cache local, só runas/itens/habilidades/feitiços, sempre com crédito e link.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const FLBLive = false
const flbAPI = "https://funleaguebuilds.de/api/wiki/guides"

var flbUA = "CienciaLOL/" + AppVersion + " (private app for 3 friends; one guide at a time, on user request)"

var flbIDRx = regexp.MustCompile(`(?:[?&]id=|/guides?/)([A-Za-z0-9_-]{6,64})`)
var flbBareRx = regexp.MustCompile(`^[A-Za-z0-9_-]{6,64}$`)

// flbGuideID aceita o link da guia (funleaguebuilds.de/guide.html?id=...) ou só o id.
func flbGuideID(s string) (string, error) {
	s = strings.TrimSpace(s)
	if flbBareRx.MatchString(s) {
		return s, nil
	}
	if !strings.Contains(s, "funleaguebuilds.de") {
		return "", fmt.Errorf("esse link não é do FunLeagueBuilds")
	}
	m := flbIDRx.FindStringSubmatch(s)
	if m == nil {
		return "", fmt.Errorf("não achei o código da guia nesse link. Copie o link da página da guia (ele tem \"?id=\" no final)")
	}
	return m[1], nil
}

// flbGuide busca UMA guia, quando o jogador cola o link dela (igual a abrir a página no navegador).
func flbGuide(link string) (string, error) {
	return flbGuideFrom(flbAPI, link)
}

func flbGuideFrom(api, link string) (string, error) {
	id, err := flbGuideID(link)
	if err != nil {
		return "", err
	}
	cl := &http.Client{Timeout: 20 * time.Second}
	req, _ := http.NewRequest("GET", api+"/"+id, nil)
	req.Header.Set("User-Agent", flbUA)
	req.Header.Set("Accept", "application/json")
	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("sem conexão com o FunLeagueBuilds (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return "", fmt.Errorf("essa guia não existe mais no FunLeagueBuilds")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("o FunLeagueBuilds respondeu com erro %d, tente de novo daqui a pouco", resp.StatusCode)
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if !json.Valid(b) {
		return "", fmt.Errorf("o FunLeagueBuilds não devolveu uma guia válida")
	}
	return string(b), nil
}

type flbCache struct {
	Fetched time.Time         `json:"fetched"`
	Guides  []json.RawMessage `json:"guides"`
}

func flbGuides(dir string, force bool) (string, error) {
	if !FLBLive {
		return "", nil // modo exemplo
	}
	path := filepath.Join(dir, "funleague.json")
	var c flbCache
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &c)
	}
	// promessa feita ao site: no máximo 1 atualização por dia, mesmo se pedirem "atualizar"
	_ = force
	if len(c.Guides) > 0 && time.Since(c.Fetched) < 24*time.Hour {
		b, _ := json.Marshal(c.Guides)
		return string(b), nil
	}
	cl := &http.Client{Timeout: 30 * time.Second}
	var all []json.RawMessage
	for page := 1; page <= 5; page++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s?limit=100&page=%d", flbAPI, page), nil)
		req.Header.Set("User-Agent", flbUA)
		resp, err := cl.Do(req)
		if err != nil {
			break
		}
		var r struct {
			Total  int               `json:"total"`
			Guides []json.RawMessage `json:"guides"`
		}
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 30<<20))
		resp.Body.Close()
		if resp.StatusCode != 200 || json.Unmarshal(b, &r) != nil {
			break
		}
		all = append(all, r.Guides...)
		if len(all) >= r.Total || len(r.Guides) == 0 {
			break
		}
	}
	if len(all) == 0 {
		if len(c.Guides) > 0 { // sem internet: usa o cache antigo
			b, _ := json.Marshal(c.Guides)
			return string(b), nil
		}
		return "", fmt.Errorf("não consegui buscar as guias do FunLeagueBuilds")
	}
	c = flbCache{Fetched: time.Now(), Guides: all}
	if b, err := json.Marshal(c); err == nil {
		_ = os.WriteFile(path, b, 0o644)
	}
	b, _ := json.Marshal(all)
	return string(b), nil
}
