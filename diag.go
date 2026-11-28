package main

// Registro de diagnóstico do automático (ban/pick) num arquivo, pra investigar quando algo não
// funciona: AppData\CienciaLoL\celular-diag.log. Não guarda nomes de jogadores.

import (
	"encoding/json"
)

// foto da seleção (só ações, tempo e a minha posição; sem os times)
func (r *remoteApp) diagSession() {
	var raw map[string]any
	if err := r.getJSON("/lol-champ-select/v1/session", &raw); err != nil {
		diag("sessão: erro %v", err)
		return
	}
	keep := map[string]any{}
	for _, k := range []string{"actions", "timer", "localPlayerCellId", "isCustomGame", "allowRerolling", "benchEnabled", "queueId"} {
		if v, ok := raw[k]; ok {
			keep[k] = v
		}
	}
	b, _ := json.Marshal(keep)
	if len(b) > 12000 {
		b = b[:12000]
	}
	diag("sessão: %s", b)
}
