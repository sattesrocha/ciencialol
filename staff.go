package main

// Trava da biblioteca: só a staff (quem tem a chave) publica ciências e correções na nuvem.
// Tudo que a staff publica vai assinado (Ed25519); os apps só aceitam da nuvem o que tem
// assinatura válida. A chave pública fica aqui (pode ser pública); a chave da staff não fica
// no código: é digitada no app por quem é da staff (ou vem embutida só na versão ADMIN).

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

var staffPub, _ = hex.DecodeString("7889098cd40574f081e5c458f5b51089c28bea4952f049afcfe21d993e58cef9")

// preenchida só na versão ADMIN (arquivo privado)
var adminStaffCode = ""

func normStaffCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), " ", ""))
}

func staffKey(code string) (ed25519.PrivateKey, error) {
	if code == "" {
		code = adminStaffCode
	}
	if code == "" {
		return nil, errors.New("sem chave da staff")
	}
	seed := sha256.Sum256([]byte("ciencialol-staff-v1:" + normStaffCode(code)))
	k := ed25519.NewKeyFromSeed(seed[:])
	if !k.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(staffPub)) {
		return nil, errors.New("chave da staff errada")
	}
	return k, nil
}

func staffCheck(code string) bool { _, err := staffKey(code); return err == nil }

func staffSign(code, data string) (string, error) {
	k, err := staffKey(code)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ed25519.Sign(k, []byte(data))), nil
}

// confere várias assinaturas de uma vez: [{"d":"...","s":"..."}] -> [true,false,...]
func staffVerify(list string) ([]bool, error) {
	var recs []struct {
		D string `json:"d"`
		S string `json:"s"`
	}
	if err := json.Unmarshal([]byte(list), &recs); err != nil {
		return nil, errors.New("lista inválida")
	}
	if len(recs) > 5000 {
		recs = recs[:5000]
	}
	out := make([]bool, len(recs))
	for i, r := range recs {
		sig, err := base64.StdEncoding.DecodeString(r.S)
		if err != nil || len(sig) != ed25519.SignatureSize || len(r.D) > 200000 {
			continue
		}
		out[i] = ed25519.Verify(ed25519.PublicKey(staffPub), []byte(r.D), sig)
	}
	return out, nil
}
