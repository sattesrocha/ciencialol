package main

// "Minhas partidas": a build que a pessoa realmente usou nas últimas partidas, lida do
// histórico do próprio cliente do LoL (runas completas, feitiços e itens).

import (
	"encoding/json"
	"time"
)

// extBuild é o formato único das builds de fora (cliente do LoL e op.gg) que a tela mostra/importa.
type extBuild struct {
	Src      string   `json:"src"` // lol | opgg | opggrec
	GameID   int64    `json:"gid,omitempty"`
	At       int64    `json:"at,omitempty"` // ms
	ChampKey int      `json:"champKey,omitempty"`
	Champ    string   `json:"champ,omitempty"`
	Win      *bool    `json:"win,omitempty"`
	K        int      `json:"k"`
	D        int      `json:"d"`
	A        int      `json:"a"`
	Queue    int      `json:"q,omitempty"`
	Mode     string   `json:"mode,omitempty"`
	Pos      string   `json:"pos,omitempty"`
	Label    string   `json:"label,omitempty"`
	Items    []int    `json:"items"`           // itens finais (ou núcleo da build recomendada)
	Start    []int    `json:"start,omitempty"` // iniciais (op.gg recomendada)
	Boots    []int    `json:"boots,omitempty"`
	Spells   []int    `json:"spells,omitempty"`
	Primary  int      `json:"primary,omitempty"`
	Sub      int      `json:"sub,omitempty"`
	Perks    []int    `json:"perks,omitempty"` // 6 runas + 3 fragmentos (pode vir só a principal)
	Rate     string   `json:"rate,omitempty"`
	Play     int      `json:"play,omitempty"`   // em quantas partidas o op.gg viu essa build
	Alt      []string `json:"alt,omitempty"`    // funções em que o op.gg tem dados do campeão
	Skills   []string `json:"skills,omitempty"` // ordem de habilidades (op.gg), ex.: Q W E Q Q R …
}

type lcuMHFull struct {
	GameID       int64  `json:"gameId"`
	GameCreation int64  `json:"gameCreation"`
	QueueID      int    `json:"queueId"`
	GameMode     string `json:"gameMode"`
	Participants []struct {
		ParticipantID int                        `json:"participantId"`
		ChampionID    int                        `json:"championId"`
		Spell1        int                        `json:"spell1Id"`
		Spell2        int                        `json:"spell2Id"`
		Stats         map[string]json.RawMessage `json:"stats"`
		Timeline      struct {
			Lane string `json:"lane"`
			Role string `json:"role"`
		} `json:"timeline"`
	} `json:"participants"`
	ParticipantIdentities []struct {
		ParticipantID int `json:"participantId"`
		Player        struct {
			Puuid string `json:"puuid"`
		} `json:"player"`
	} `json:"participantIdentities"`
}

func statInt(m map[string]json.RawMessage, k string) int {
	var v int
	if r, ok := m[k]; ok {
		_ = json.Unmarshal(r, &v)
	}
	return v
}

func statBool(m map[string]json.RawMessage, k string) bool {
	var v bool
	if r, ok := m[k]; ok {
		_ = json.Unmarshal(r, &v)
	}
	return v
}

func lcuMyBuilds() ([]extBuild, error) {
	c, err := lcuConnect()
	if err != nil {
		return nil, err
	}
	c.cl.Timeout = 45 * time.Second
	c.deadline = time.Now().Add(60 * time.Second)
	return lcuMyBuildsWith(c)
}

func lcuMyBuildsWith(c *lcuConn) ([]extBuild, error) {
	var me struct {
		Puuid string `json:"puuid"`
	}
	if _, err := c.do("GET", "/lol-summoner/v1/current-summoner", nil, &me); err != nil {
		return nil, err
	}
	var mh struct {
		Games struct {
			Games []lcuMHFull `json:"games"`
		} `json:"games"`
	}
	_, err := c.do("GET", "/lol-match-history/v1/products/lol/current-summoner/matches?begIndex=0&endIndex=20", nil, &mh)
	if err != nil && me.Puuid != "" {
		_, err = c.do("GET", "/lol-match-history/v1/products/lol/"+me.Puuid+"/matches?begIndex=0&endIndex=20", nil, &mh)
	}
	if err != nil {
		return nil, err
	}
	out := []extBuild{}
	for _, g := range mh.Games.Games {
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
			win := statBool(s, "win")
			b := extBuild{Src: "lol", GameID: g.GameID, At: g.GameCreation, ChampKey: p.ChampionID, Win: &win,
				K: statInt(s, "kills"), D: statInt(s, "deaths"), A: statInt(s, "assists"), Queue: g.QueueID, Mode: g.GameMode,
				Pos: p.Timeline.Lane + "/" + p.Timeline.Role, Items: []int{}}
			for i := 0; i <= 6; i++ {
				if v := statInt(s, "item"+string(rune('0'+i))); v > 0 {
					b.Items = append(b.Items, v)
				}
			}
			if p.Spell1 > 0 && p.Spell2 > 0 {
				b.Spells = []int{p.Spell1, p.Spell2}
			}
			b.Primary, b.Sub = statInt(s, "perkPrimaryStyle"), statInt(s, "perkSubStyle")
			for i := 0; i < 6; i++ {
				if v := statInt(s, "perk"+string(rune('0'+i))); v > 0 {
					b.Perks = append(b.Perks, v)
				}
			}
			sh := []int{statInt(s, "statPerk0"), statInt(s, "statPerk1"), statInt(s, "statPerk2")}
			if len(b.Perks) == 6 && sh[0] > 0 && sh[1] > 0 && sh[2] > 0 {
				b.Perks = append(b.Perks, sh...)
			}
			out = append(out, b)
			break
		}
	}
	return out, nil
}
