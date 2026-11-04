package main

// Senha mestre das salas: cada sala guarda um "envelope" com a senha dela, trancado com a
// chave pública abaixo. Só quem sabe a senha mestre (que não está em lugar nenhum do app)
// consegue abrir o envelope e entrar em qualquer sala.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

const masterPub = "T09TDul0Lgd8j+oYkSBSAGmtiajsQJ739VnXdjj2fno="

const masterSalt = "CienciaLOL-mestre-v1"
const masterIter = 200000

func envKey(shared, eph, pub []byte) []byte {
	h := sha256.New()
	h.Write([]byte("CienciaLOL-envelope-v1"))
	h.Write(shared)
	h.Write(eph)
	h.Write(pub)
	return h.Sum(nil)
}

// roomSeal tranca um texto pra só a senha mestre abrir.
func roomSeal(plain string) (string, error) {
	pubB, err := base64.StdEncoding.DecodeString(masterPub)
	if err != nil {
		return "", err
	}
	pub, err := ecdh.X25519().NewPublicKey(pubB)
	if err != nil {
		return "", err
	}
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	shared, err := eph.ECDH(pub)
	if err != nil {
		return "", err
	}
	blk, _ := aes.NewCipher(envKey(shared, eph.PublicKey().Bytes(), pubB))
	gcm, _ := cipher.NewGCM(blk)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	out := append(append(eph.PublicKey().Bytes(), nonce...), gcm.Seal(nil, nonce, []byte(plain), nil)...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// roomOpen abre o envelope com a senha mestre (erro se a senha não for a mestre).
func roomOpen(sealed, password string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil || len(b) < 32+12+16 {
		return "", errors.New("envelope inválido")
	}
	k, err := pbkdf2.Key(sha256.New, password, []byte(masterSalt), masterIter, 32)
	if err != nil {
		return "", err
	}
	priv, err := ecdh.X25519().NewPrivateKey(k)
	if err != nil {
		return "", err
	}
	pubB := priv.PublicKey().Bytes()
	if base64.StdEncoding.EncodeToString(pubB) != masterPub {
		return "", errors.New("senha incorreta")
	}
	eph, err := ecdh.X25519().NewPublicKey(b[:32])
	if err != nil {
		return "", err
	}
	shared, err := priv.ECDH(eph)
	if err != nil {
		return "", err
	}
	blk, _ := aes.NewCipher(envKey(shared, b[:32], pubB))
	gcm, _ := cipher.NewGCM(blk)
	plain, err := gcm.Open(nil, b[32:44], b[44:], nil)
	if err != nil {
		return "", errors.New("envelope inválido")
	}
	return string(plain), nil
}
