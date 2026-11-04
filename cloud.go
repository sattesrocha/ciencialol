package main

// Nuvem no Drive: um Google Apps Script do dono do app guarda a biblioteca num
// arquivo JSON do Drive dele. O app busca tudo (GET) e manda a sua parte (POST).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Endereço e chave da nuvem: ficam em private_config.go, que não vai pra versão pública.
// Sem eles, a nuvem fica desligada (o app funciona normal, só não sincroniza).
var (
	CloudURL = ""
	CloudKey = ""
)

var cloudKeyRx = regexp.MustCompile(`^[mp]/[\w-]{1,40}$`)

type cloudResp struct {
	Ok    bool            `json:"ok"`
	Erro  string          `json:"erro"`
	Dados json.RawMessage `json:"dados"`
}

func cloudClient() *http.Client { return &http.Client{Timeout: 40 * time.Second} }

func cloudRead(resp *http.Response) (cloudResp, error) {
	var r cloudResp
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return r, fmt.Errorf("o Google respondeu com erro %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal(b, &r); err != nil {
		// mostra um trecho do que o Google devolveu (sem as tags), pra saber o motivo
		txt := regexp.MustCompile(`(?s)<script.*?</script>|<style.*?</style>|<[^>]+>`).ReplaceAllString(string(b), " ")
		txt = strings.Join(strings.Fields(txt), " ")
		if len(txt) > 300 {
			txt = txt[:300]
		}
		return r, fmt.Errorf("o script do Drive não respondeu direito. Resposta do Google: %q", txt)
	}
	if !r.Ok {
		return r, errors.New("o script do Drive recusou: " + r.Erro)
	}
	return r, nil
}

// cloudGet devolve o JSON com tudo que está guardado ({"chave":{"ts":..,"dados":".."}}).
func cloudGet() (string, error) { return cloudGetFrom(CloudURL) }

func cloudGetFrom(base string) (string, error) {
	resp, err := cloudClient().Get(base + "?k=" + url.QueryEscape(CloudKey))
	if err != nil {
		return "", fmt.Errorf("sem conexão com o Drive (%v)", err)
	}
	r, err := cloudRead(resp)
	if err != nil {
		return "", err
	}
	if len(r.Dados) == 0 {
		return "{}", nil
	}
	return string(r.Dados), nil
}

// cloudPutAdmin grava a configuração do grupo (c/...). O script só aceita com a senha de admin,
// que existe apenas na versão admin do app (arquivo admin_windows.go, compilado com a tag "admin").
func cloudPutAdminTo(base, adminKey, key, data string) error {
	if !regexp.MustCompile(`^c/[\w-]{1,40}$`).MatchString(key) || len(data) > 1000000 || adminKey == "" {
		return errors.New("dados inválidos")
	}
	body, _ := json.Marshal(map[string]string{"k": CloudKey, "admin": adminKey, "chave": key, "dados": data})
	resp, err := cloudClient().Post(base, "text/plain;charset=utf-8", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sem conexão com o Drive (%v)", err)
	}
	_, err = cloudRead(resp)
	return err
}

// cloudPut grava a parte deste app (m/<id> = biblioteca, p/<nome> = backup pessoal).
func cloudPut(key, data string) error { return cloudPutTo(CloudURL, key, data) }

func cloudPutTo(base, key, data string) error {
	if !cloudKeyRx.MatchString(key) || len(data) > 2000000 {
		return errors.New("dados inválidos")
	}
	body, _ := json.Marshal(map[string]string{"k": CloudKey, "chave": key, "dados": data})
	resp, err := cloudClient().Post(base, "text/plain;charset=utf-8", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sem conexão com o Drive (%v)", err)
	}
	_, err = cloudRead(resp)
	return err
}
