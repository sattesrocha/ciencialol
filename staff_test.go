package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"testing"
)

func TestStaffSignVerify(t *testing.T) {
	// chave de teste (a de verdade não fica no código)
	code := "CLOL-TEST-TEST-TEST-TEST"
	seed := sha256.Sum256([]byte("ciencialol-staff-v1:" + code))
	old := staffPub
	staffPub = ed25519.NewKeyFromSeed(seed[:]).Public().(ed25519.PublicKey)
	defer func() { staffPub = old }()

	if !staffCheck(code) || !staffCheck(" clol-test-test-test-test ") {
		t.Fatal("a chave da staff devia valer")
	}
	if staffCheck("CLOL-AAAA-BBBB-CCCC-DDDD") {
		t.Fatal("chave errada passou")
	}
	s, err := staffSign(code, `{"t":"c","id":"custom-1"}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := staffSign("errada", "x"); err == nil {
		t.Fatal("assinou com chave errada")
	}
	in, _ := json.Marshal([]map[string]string{{"d": `{"t":"c","id":"custom-1"}`, "s": s}, {"d": `{"t":"c","id":"custom-2"}`, "s": s}, {"d": "x", "s": "lixo"}})
	ok, err := staffVerify(string(in))
	if err != nil || len(ok) != 3 || !ok[0] || ok[1] || ok[2] {
		t.Fatalf("verificação errada: %v %v", ok, err)
	}
}
