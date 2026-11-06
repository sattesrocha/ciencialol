package main

import (
	"errors"
	"strconv"
	"time"
)

// Partida completa (os 10 jogadores), igual à tela de detalhes do histórico do LoL.
type mgPlayer struct {
	Name   string `json:"name"`
	Champ  int    `json:"champ"`
	Spells [2]int `json:"spells"`
	Items  []int  `json:"items"`
	Rune   int    `json:"rune"` // runa principal (keystone)
	Sub    int    `json:"sub"`  // árvore secundária
	K      int    `json:"k"`
	D      int    `json:"d"`
	A      int    `json:"a"`
	CS     int    `json:"cs"`
	Dmg    int    `json:"dmg"`
	Gold   int    `json:"gold"`
	Vis    int    `json:"vis"`
	Lvl    int    `json:"lvl"`
}
type mgTeam struct {
	ID      int        `json:"id"`
	Win     bool       `json:"win"`
	Players []mgPlayer `json:"players"`
}
type mgGame struct {
	GameID   int64    `json:"gid"`
	Creation int64    `json:"creation"`
	Duration int      `json:"dur"`
	Mode     string   `json:"mode"`
	Queue    int      `json:"queue"`
	Teams    []mgTeam `json:"teams"`
}

type lcuFullGame struct {
	GameID       int64  `json:"gameId"`
	GameCreation int64  `json:"gameCreation"`
	GameDuration int    `json:"gameDuration"`
	QueueID      int    `json:"queueId"`
	GameMode     string `json:"gameMode"`
	Participants []struct {
		ParticipantID int `json:"participantId"`
		TeamID        int `json:"teamId"`
		ChampionID    int `json:"championId"`
		Spell1ID      int `json:"spell1Id"`
		Spell2ID      int `json:"spell2Id"`
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
			Perk0                       int  `json:"perk0"`
			PerkSubStyle                int  `json:"perkSubStyle"`
		} `json:"stats"`
	} `json:"participants"`
	ParticipantIdentities []struct {
		ParticipantID int `json:"participantId"`
		Player        struct {
			GameName     string `json:"gameName"`
			TagLine      string `json:"tagLine"`
			SummonerName string `json:"summonerName"`
		} `json:"player"`
	} `json:"participantIdentities"`
}

func toMgGame(g lcuFullGame) mgGame {
	names := map[int]string{}
	for _, id := range g.ParticipantIdentities {
		n := id.Player.GameName
		if n == "" {
			n = id.Player.SummonerName
		}
		names[id.ParticipantID] = n
	}
	out := mgGame{GameID: g.GameID, Creation: g.GameCreation, Duration: g.GameDuration, Mode: g.GameMode, Queue: g.QueueID}
	idx := map[int]int{}
	for _, p := range g.Participants {
		ti, ok := idx[p.TeamID]
		if !ok {
			ti = len(out.Teams)
			idx[p.TeamID] = ti
			out.Teams = append(out.Teams, mgTeam{ID: p.TeamID, Win: p.Stats.Win})
		}
		s := p.Stats
		out.Teams[ti].Players = append(out.Teams[ti].Players, mgPlayer{
			Name: names[p.ParticipantID], Champ: p.ChampionID, Spells: [2]int{p.Spell1ID, p.Spell2ID},
			Items: []int{s.Item0, s.Item1, s.Item2, s.Item3, s.Item4, s.Item5, s.Item6}, Rune: s.Perk0, Sub: s.PerkSubStyle,
			K: s.Kills, D: s.Deaths, A: s.Assists, CS: s.TotalMinionsKilled + s.NeutralMinionsKilled,
			Dmg: s.TotalDamageDealtToChampions, Gold: s.GoldEarned, Vis: s.VisionScore, Lvl: s.ChampLevel,
		})
	}
	return out
}

func lcuGame(gid string) (mgGame, error) {
	id, err := strconv.ParseInt(gid, 10, 64)
	if err != nil || id <= 0 {
		return mgGame{}, errors.New("partida sem número do LoL")
	}
	c, err := lcuConnect()
	if err != nil {
		return mgGame{}, err
	}
	c.cl.Timeout = 30 * time.Second // o cliente busca a partida nos servidores da Riot
	defer c.cl.CloseIdleConnections()
	return lcuGameWith(c, id)
}

func lcuGameWith(c *lcuConn, id int64) (mgGame, error) {
	var g lcuFullGame
	st, err := c.do("GET", "/lol-match-history/v1/games/"+strconv.FormatInt(id, 10), nil, &g)
	if err != nil {
		if st == 404 {
			return mgGame{}, errors.New("o LoL não achou essa partida (ela pode ser de outra conta ou antiga demais)")
		}
		return mgGame{}, err
	}
	if len(g.Participants) == 0 {
		return mgGame{}, errors.New("o LoL não devolveu os jogadores dessa partida")
	}
	return toMgGame(g), nil
}
