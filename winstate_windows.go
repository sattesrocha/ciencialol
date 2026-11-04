//go:build windows && !semjanela

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

var (
	u32dll         = windows.NewLazySystemDLL("user32.dll")
	procIsIconic   = u32dll.NewProc("IsIconic")
	procShowWindow = u32dll.NewProc("ShowWindow")
	procSetFG      = u32dll.NewProc("SetForegroundWindow")
	procBringTop   = u32dll.NewProc("BringWindowToTop")
	procSetWinPos  = u32dll.NewProc("SetWindowPos")
	procGetWTP     = u32dll.NewProc("GetWindowThreadProcessId")
	procAttach     = u32dll.NewProc("AttachThreadInput")
	procFlashWinEx = u32dll.NewProc("FlashWindowEx")
	procSysParamsW = u32dll.NewProc("SystemParametersInfoW")
	procGetFG      = u32dll.NewProc("GetForegroundWindow")
)

type flashInfo struct {
	cbSize    uint32
	hwnd      uintptr
	dwFlags   uint32
	uCount    uint32
	dwTimeout uint32
}

// traz a janela do app pra frente (mesmo com outro programa, inclusive o LoL, em foco)
func focusWindow(h uintptr) {
	if h == 0 {
		return
	}
	if r, _, _ := procIsIconic.Call(h); r != 0 {
		procShowWindow.Call(h, 9) // SW_RESTORE
	}
	// o Windows só deixa trocar o foco pra quem está ligado à janela da frente: liga as duas por um instante
	fg, _, _ := procGetFG.Call()
	if fg != 0 && fg != h {
		fgThread, _, _ := procGetWTP.Call(fg, 0)
		me := uintptr(windows.GetCurrentThreadId())
		if fgThread != 0 && fgThread != me {
			procAttach.Call(me, fgThread, 1)
			defer procAttach.Call(me, fgThread, 0)
		}
	}
	hwndTopmost, hwndNoTopmost := ^uintptr(0), ^uintptr(1) // -1, -2
	procSetWinPos.Call(h, hwndTopmost, 0, 0, 0, 0, 0x0003|0x0040)
	procSetWinPos.Call(h, hwndNoTopmost, 0, 0, 0, 0, 0x0003|0x0040)
	procSetFG.Call(h)
	procBringTop.Call(h)
	fi := flashInfo{hwnd: h, dwFlags: 0x3 | 0xC, uCount: 3} // pisca na barra de tarefas até ganhar foco
	fi.cbSize = uint32(unsafe.Sizeof(fi))
	procFlashWinEx.Call(uintptr(unsafe.Pointer(&fi)))
}

func init() {
	extraBinds = append(extraBinds, func(w webview2.WebView) {
		{
			logStep("janela: binds")
			w.Bind("focusApp", func() (ok bool) {
				defer func() {
					if r := recover(); r != nil {
						logStep(fmt.Sprintf("focusApp falhou: %v", r))
						ok = false
					}
				}()
				h := uintptr(w.Window())
				focusWindow(h)
				r, _, _ := procGetFG.Call()
				return r == h
			})
			// papel de parede: copia o arquivo que vem dentro do app pra pasta do app e define no Windows
			w.Bind("wallpaperSet", func(name string) (res string, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("erro interno: %v", r)
					}
				}()
				p, err := wallFile(name)
				if err != nil {
					return "", err
				}
				ptr, _ := windows.UTF16PtrFromString(p)
				r, _, e := procSysParamsW.Call(0x0014, 0, uintptr(unsafe.Pointer(ptr)), 0x01|0x02) // SPI_SETDESKWALLPAPER
				if r == 0 {
					return "", e
				}
				return p, nil
			})
			// salva na pasta Imagens\Ciência LOL e abre o Explorer com o arquivo selecionado
			w.Bind("wallpaperSave", func(name string) (string, error) {
				b, err := uiFiles.ReadFile("ui/wall/" + filepath.Base(name))
				if err != nil {
					return "", err
				}
				home, _ := os.UserHomeDir()
				dir := filepath.Join(home, "Pictures", "Ciência LOL")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return "", err
				}
				p := filepath.Join(dir, filepath.Base(name))
				if err := os.WriteFile(p, b, 0o644); err != nil {
					return "", err
				}
				exec.Command("explorer", "/select,", p).Start()
				return p, nil
			})
		}
		logStep("janela: binds ok")
		// avisa a página quando a janela é minimizada ou volta (pra música e o "welcome back")
		go func() {
			defer func() { recover() }()
			last := ""
			for {
				time.Sleep(400 * time.Millisecond)
				h := uintptr(w.Window())
				if h == 0 {
					continue
				}
				r, _, _ := procIsIconic.Call(h)
				st := "normal"
				if r != 0 {
					st = "min"
				}
				if st != last {
					if last == "" {
						logStep("janela: estado " + st)
					}
					last = st
					s := st
					w.Dispatch(func() { w.Eval("window.__winState && window.__winState('" + s + "')") })
				}
			}
		}()
	})
}

func wallFile(name string) (string, error) {
	b, err := uiFiles.ReadFile("ui/wall/" + filepath.Base(name))
	if err != nil {
		return "", err
	}
	dir := filepath.Join(appDir(), "wallpaper")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, filepath.Base(name))
	return p, os.WriteFile(p, b, 0o644)
}
