package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// garante que toda DLL do Windows usada pelo app tem nome de verdade (um nome errado derruba o programa)
func TestDLLNames(t *testing.T) {
	ok := map[string]bool{"user32.dll": true, "winmm.dll": true, "kernel32.dll": true, "shell32.dll": true}
	rx := regexp.MustCompile(`NewLazySystemDLL\("([^"]+)"\)`)
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		b, _ := os.ReadFile(f)
		for _, m := range rx.FindAllStringSubmatch(string(b), -1) {
			if !ok[m[1]] {
				t.Errorf("%s: DLL desconhecida %q", f, m[1])
			}
		}
	}
}
