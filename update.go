package main

// Atualização automática: o programa baixa o .exe publicado no Google Drive
// (link fixo, "qualquer pessoa com o link"), lê a versão gravada dentro dele
// e, se for mais nova, se substitui. Arquivos baixados pelo próprio programa
// não recebem a marca "veio da internet", então o Windows não mostra o aviso.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// AppVersion é a versão deste executável. Suba a cada versão publicada.
const AppVersion = "2.79"

// ManifestDocID é um Google Doc fixo ("qualquer pessoa com o link") com duas linhas:
//
//	versao=2.5.1
//	arquivo=<ID do CienciaLOL.exe no Drive>
//
// Se o .exe for trocado e ganhar outro link, basta atualizar o documento.
const ManifestDocID = "1Buw2saZwYVAxl3BkaRDWtE9W3m1ga9aWAwo1MpU8gW8"

var (
	manVerRx  = regexp.MustCompile(`(?m)^\s*versao\s*=\s*(\d+(?:\.\d+)*)\s*$`)
	manFileRx = regexp.MustCompile(`(?m)^\s*arquivo\s*=\s*([\w-]{20,})\s*$`)
)

func manifestURL(id string) string {
	return "https://docs.google.com/document/d/" + id + "/export?format=txt"
}

// fetchManifest lê o documento de atualização e devolve versão e ID do .exe.
func fetchManifest(url string) (string, string, error) {
	cl := &http.Client{Timeout: 30 * time.Second}
	resp, err := cl.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("sem conexão com o Google Drive (%v)", err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == 404:
		return "", "", errors.New("o documento de atualização não foi encontrado no Google Drive")
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return "", "", errors.New("o documento de atualização não está liberado (precisa ser \"Qualquer pessoa com o link\")")
	case resp.StatusCode != 200:
		return "", "", fmt.Errorf("o Google Drive respondeu com erro %d, tente de novo em alguns minutos", resp.StatusCode)
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
	txt := strings.ReplaceAll(strings.TrimPrefix(string(b), "\ufeff"), "\r", "")
	v, f := manVerRx.FindStringSubmatch(txt), manFileRx.FindStringSubmatch(txt)
	if v == nil || f == nil {
		return "", "", errors.New("não consegui ler o documento de atualização (ele está compartilhado como \"Qualquer pessoa com o link\"?)")
	}
	return v[1], f[1], nil
}

// O marcador fica gravado dentro do .exe; é assim que uma versão reconhece a outra.
var versionTag = "CIENCIALOL_VERSION=" + AppVersion + ";"

var markerRx = regexp.MustCompile(`CIENCIALOL_VERSION=(\d+(?:\.\d+)*);`)

// versionIn procura o marcador de versão dentro de um executável baixado.
func versionIn(b []byte) string {
	for _, m := range markerRx.FindAllSubmatch(b, -1) {
		return string(m[1])
	}
	return ""
}

// newer diz se a versão a é maior que b (comparação numérica por partes).
func newer(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}

func driveURL(id string) string {
	return "https://drive.usercontent.google.com/download?id=" + id + "&export=download&confirm=t"
}

// download baixa o .exe publicado e valida que é mesmo um executável do Ciência LOL.
func download(url string) ([]byte, string, error) {
	cl := &http.Client{Timeout: 90 * time.Second}
	resp, err := cl.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("sem conexão com o Google Drive (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, "", errors.New("o arquivo do programa não foi encontrado no Google Drive (ele pode ter sido apagado ou substituído por outro, em vez de receber uma nova versão)")
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, "", errors.New("o arquivo no Google Drive não está liberado (o compartilhamento precisa ser \"Qualquer pessoa com o link\")")
	}
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("o Google Drive respondeu com erro %d, tente de novo em alguns minutos", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 80<<20))
	if err != nil {
		return nil, "", fmt.Errorf("download interrompido (%v)", err)
	}
	if len(b) < 1<<20 || !bytes.HasPrefix(b, []byte("MZ")) {
		return nil, "", errors.New("o Google Drive não entregou o programa. Confira se o compartilhamento é \"Qualquer pessoa com o link\"")
	}
	v := versionIn(b)
	if v == "" {
		return nil, "", errors.New("o arquivo no Google Drive é de uma versão antiga do programa, sem atualização automática")
	}
	return b, v, nil
}

// estado entre "verificar" e "aplicar"
var pending struct {
	data    []byte
	version string
}

type updateInfo struct {
	Status  string `json:"status"` // none | available | error | off
	Current string `json:"current"`
	Version string `json:"version,omitempty"`
	Message string `json:"message,omitempty"`
}

// peekUpdate só lê o documento de atualização (rápido, sem baixar o .exe)
// e diz se existe uma versão mais nova. Usado pro aviso no botão.
func peekUpdate() updateInfo {
	return peekUpdateFrom(manifestURL(ManifestDocID))
}

func peekUpdateFrom(manURL string) updateInfo {
	info := updateInfo{Current: versionIn([]byte(versionTag))}
	ver, _, err := fetchManifest(manURL)
	if err != nil {
		info.Status, info.Message = "error", err.Error()
		return info
	}
	info.Version = ver
	if newer(ver, info.Current) {
		info.Status = "available"
	} else {
		info.Status = "none"
	}
	return info
}

func checkUpdate() updateInfo {
	return checkUpdateFrom(manifestURL(ManifestDocID), driveURL)
}

func checkUpdateFrom(manURL string, fileURL func(string) string) updateInfo {
	info := updateInfo{Current: versionIn([]byte(versionTag))}
	ver, fileID, err := fetchManifest(manURL)
	if err != nil {
		info.Status, info.Message = "error", err.Error()
		return info
	}
	if !newer(ver, info.Current) {
		info.Status, info.Version = "none", ver
		return info
	}
	b, got, err := download(fileURL(fileID))
	if err != nil {
		info.Status, info.Message = "error", err.Error()
		return info
	}
	if !newer(got, info.Current) {
		info.Status, info.Message = "error", fmt.Sprintf("o documento anuncia a versão %s, mas o arquivo no Drive ainda é a %s", ver, got)
		return info
	}
	pending.data, pending.version = b, got
	info.Status, info.Version = "available", got
	return info
}

// stageUpdate grava a versão nova no lugar da atual, com o mesmo nome de arquivo
// (ex.: CienciaLOL.exe). A atual é renomeada para .old e apagada pela nova.
func stageUpdate(exe string) (string, error) {
	if pending.data == nil {
		return "", errors.New("nenhuma atualização baixada")
	}
	tmp := exe + ".download"
	if err := os.WriteFile(tmp, pending.data, 0o755); err != nil {
		return "", fmt.Errorf("não consegui gravar na pasta do programa (%v). Mova o programa para outra pasta, como a Área de Trabalho, e tente de novo", err)
	}
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("não consegui substituir o programa (%v)", err)
	}
	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Rename(old, exe) // desfaz
		return "", fmt.Errorf("não consegui finalizar a atualização (%v)", err)
	}
	return exe, nil
}

// cleanupOld apaga o executável antigo depois que a versão nova abriu.
func cleanupOld(path string) {
	if !strings.HasSuffix(strings.ToLower(path), ".exe.old") {
		return
	}
	for i := 0; i < 40; i++ {
		if err := os.Remove(path); err == nil || os.IsNotExist(err) {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}
