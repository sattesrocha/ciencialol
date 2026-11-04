//go:build windows

package main

// Sons tocados pelo próprio programa (e não pelo navegador interno), pra que o
// Mixer de Volume do Windows mostre "Ciência LOL" com o ícone do app em vez de
// "Microsoft Edge WebView2". Usa o MCI do Windows (winmm.dll), que toca mp3.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

var (
	winmm   = windows.NewLazySystemDLL("winmm.dll")
	procMci = winmm.NewProc("mciSendStringW")
)

func mci(cmd string) (string, error) {
	p, _ := syscall.UTF16PtrFromString(cmd)
	buf := make([]uint16, 128)
	r, _, _ := procMci.Call(uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
	if r != 0 {
		return "", fmt.Errorf("mci %d", r)
	}
	return syscall.UTF16ToString(buf), nil
}

type audioSys struct {
	dir   string
	ch    chan func()
	open  map[int]bool // aliases abertos (pra fechar os que já terminaram)
	loops map[int]bool // sons em loop (música): só fecham quando mandarem parar
	ready chan bool
}

func newAudio(appDir string) *audioSys {
	a := &audioSys{dir: filepath.Join(appDir, "snd"), ch: make(chan func(), 256), open: map[int]bool{}, loops: map[int]bool{}, ready: make(chan bool, 1)}
	go a.loop()
	return a
}

// todos os comandos MCI rodam numa linha só (o MCI exige isso)
func (a *audioSys) loop() {
	runtime.LockOSThread()
	ok := a.extract() == nil
	if ok {
		// teste: abre e fecha um som
		if _, err := mci(`open "` + filepath.Join(a.dir, "sfx_blip0.mp3") + `" type mpegvideo alias probe`); err != nil {
			ok = false
		} else {
			mci("close probe")
		}
	}
	a.ready <- ok
	for f := range a.ch {
		f()
	}
}

// copia os sons de dentro do .exe pra pasta do app (o MCI precisa de arquivo)
func (a *audioSys) extract() error {
	if err := os.MkdirAll(a.dir, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(uiFiles, "ui/snd", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := uiFiles.ReadFile(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(a.dir, filepath.Base(p))
		if st, e := os.Stat(dst); e == nil && st.Size() == int64(len(b)) {
			return nil
		}
		return os.WriteFile(dst, b, 0o644)
	})
}

func (a *audioSys) do(f func()) {
	select {
	case a.ch <- f:
	default: // fila cheia: ignora (é só som)
	}
}

func clampVol(v int) int {
	if v < 0 {
		return 0
	}
	if v > 1000 {
		return 1000
	}
	return v
}

func validSound(name string) bool {
	if name == "" || len(name) > 40 {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}

func (a *audioSys) play(id int, name string, vol int, loop bool) {
	if !validSound(name) {
		return
	}
	a.do(func() {
		// fecha os que já acabaram
		for k := range a.open {
			if a.loops[k] {
				continue
			}
			if st, err := mci(fmt.Sprintf("status s%d mode", k)); err != nil || strings.HasPrefix(st, "stopped") {
				mci(fmt.Sprintf("close s%d", k))
				delete(a.open, k)
			}
		}
		al := fmt.Sprintf("s%d", id)
		if _, err := mci(`open "` + filepath.Join(a.dir, name+".mp3") + `" type mpegvideo alias ` + al); err != nil {
			return
		}
		a.open[id] = true
		if loop {
			a.loops[id] = true
		}
		mci(fmt.Sprintf("setaudio %s volume to %d", al, clampVol(vol)))
		if loop {
			mci("play " + al + " repeat")
		} else {
			mci("play " + al + " from 0")
		}
	})
}

func (a *audioSys) volume(id, vol int) {
	a.do(func() {
		if a.open[id] {
			mci(fmt.Sprintf("setaudio s%d volume to %d", id, clampVol(vol)))
		}
	})
}

func (a *audioSys) stop(id int) {
	a.do(func() {
		if a.open[id] {
			mci(fmt.Sprintf("close s%d", id))
			delete(a.open, id)
			delete(a.loops, id)
		}
	})
}

func init() {
	extraBinds = append(extraBinds, func(w webview2.WebView) {
		au := newAudio(appDir())
		var ok *bool
		w.Bind("sndReady", func() bool {
			if ok == nil {
				v := <-au.ready
				ok = &v
			}
			return *ok
		})
		w.Bind("sndPlay", func(id int, name string, vol int, loop bool) { au.play(id, name, vol, loop) })
		w.Bind("sndVol", func(id, vol int) { au.volume(id, vol) })
		w.Bind("sndStop", func(id int) { au.stop(id) })
	})
}
