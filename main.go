//go:build windows

package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

//go:embed ui
var uiFiles embed.FS

var (
	user32         = windows.NewLazySystemDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

// extraBinds: funções extras ligadas só em versões especiais (ex.: build com a tag "lcu")
var extraBinds []func(webview2.WebView)

// maximiza a janela depois que a tela carregou (definida em winstate_windows.go)

func msgBox(title, text string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(text)
	procMessageBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), 0x40)
}

func appDir() string {
	base, err := os.UserConfigDir() // %APPDATA%
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "CienciaLoL")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func main() {
	// Versão nova aberta pela atualização: apaga o executável antigo.
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "--cleanup=") {
			go cleanupOld(strings.TrimPrefix(a, "--cleanup="))
		}
	}
	dir := appDir()
	statePath := filepath.Join(dir, "dados.json")
	startLog(dir)
	logStep("início (versão " + AppVersion + ")")
	opggInit(dir)
	// se o programa der erro e fechar, guarda o motivo e avisa (em vez de sumir sem explicação)
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%s versão %s\n%v\n\n%s", time.Now().Format(time.RFC3339), AppVersion, r, debug.Stack())
			_ = os.WriteFile(filepath.Join(dir, "erro.log"), []byte(msg), 0o644)
			msgBox("Ciência LOL", fmt.Sprintf("O Ciência LOL fechou por causa de um erro:\n\n%v\n\nO detalhe ficou salvo em:\n%s", r, filepath.Join(dir, "erro.log")))
		}
	}()

	// Servidor local só para a interface (fica preso no 127.0.0.1).
	sub, _ := fs.Sub(uiFiles, "ui")
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		msgBox("Ciência LOL", "Não foi possível iniciar o programa: "+err.Error())
		return
	}
	go http.Serve(ln, http.FileServer(http.FS(sub)))
	logStep("servidor local ok")

	// Deixa os sons tocarem mesmo para quem só está assistindo a sala (sem clique antes).
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--autoplay-policy=no-user-gesture-required")

	cacheDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "CienciaLoL", "WebView2")
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		DataPath:  cacheDir,
		WindowOptions: webview2.WindowOptions{
			Title:  "Ciência LOL",
			Width:  1320,
			Height: 860,
			IconId: 2, // grupo de ícone (o manifesto ocupa o ID 1)
			Center: true,
		},
	})
	if w == nil {
		msgBox("Ciência LOL", "Este programa precisa do Microsoft Edge WebView2 Runtime (já vem no Windows 10 e 11).\n\nSe ele não abrir, instale em:\nhttps://developer.microsoft.com/microsoft-edge/webview2/")
		return
	}
	defer w.Destroy()
	// segurança: se a janela for parar em qualquer outra página (link externo, redirecionamento),
	// ela perde o acesso às funções do programa (salvar, arquivos, LoL, atualização...)
	origin := "http://" + ln.Addr().String()
	w.Init(`(function(){ if (location.origin !== ` + strconv.Quote(origin) + ` && location.href !== "about:blank") { try {
		var wv = window.chrome && window.chrome.webview; if (wv) { Object.defineProperty(wv, "postMessage", { value: function(){}, writable: false, configurable: false }); }
	} catch (e) {} } })();`)
	logStep("janela e WebView2 criados")
	w.SetSize(1100, 700, webview2.HintMin)

	w.Bind("logPage", func(s string) {
		logStep("tela: " + s)
	})
	w.Bind("loadState", func() string {
		b, err := os.ReadFile(statePath)
		if err != nil {
			return ""
		}
		return string(b)
	})
	w.Bind("saveState", func(s string) error {
		tmp := statePath + ".tmp"
		if err := os.WriteFile(tmp, []byte(s), 0o644); err != nil {
			return err
		}
		return os.Rename(tmp, statePath)
	})
	allowed := regexp.MustCompile(`^https://(www\.)?(instagram\.com|linktr\.ee|youtube\.com|tiktok\.com|funleaguebuilds\.de|op\.gg)/`)
	w.Bind("openURL", func(u string) {
		if !allowed.MatchString(u) {
			return
		}
		verb, _ := syscall.UTF16PtrFromString("open")
		target, _ := syscall.UTF16PtrFromString(u)
		windows.ShellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL)
	})
	w.Bind("exportFile", func(name, content string) (string, error) {
		name = filepath.Base(strings.ReplaceAll(name, "..", ""))
		// só deixa salvar backup .json com nome simples (nada de .exe/.bat/.lnk na Área de Trabalho)
		if !regexp.MustCompile(`^[\w .()-]{1,80}\.json$`).MatchString(name) {
			return "", errors.New("nome de arquivo inválido")
		}
		home, _ := os.UserHomeDir()
		dest := filepath.Join(home, "Desktop")
		if st, err := os.Stat(dest); err != nil || !st.IsDir() {
			dest = dir
		}
		p := filepath.Join(dest, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			return "", err
		}
		return p, nil
	})

	// funções que usam a internet: em segundo plano, pra janela não travar
	w.Init(asyncJS)
	// tema de Halloween já na abertura (sem piscar o visual normal antes)
	w.Init("window.__HWPREF=" + strconv.FormatBool(halloweenPref(statePath)) + ";")
	if roomSecret != "" {
		w.Init("window.__RS=" + strconv.Quote(roomSecret) + ";")
	}
	bindAsync(w, "opggChampBuild", func(a []string) (any, error) { return opggChampBuild(arg(a, 0), arg(a, 1), arg(a, 2)) })
	bindAsync(w, "flbGuides", func(a []string) (any, error) { return flbGuides(dir, arg(a, 0) == "true") })
	bindAsync(w, "flbGuide", func(a []string) (any, error) { return flbGuide(arg(a, 0)) })
	{ // atualização pelo Drive (manifesto no Google Doc)
		bindAsync(w, "checkUpdate", func(a []string) (any, error) { return checkUpdate(), nil })
		bindAsync(w, "peekUpdate", func(a []string) (any, error) { return peekUpdate(), nil })
	}
	// importar pro LoL, partidas do cliente e senha mestre das salas
	bindAsync(w, "lcuImport", func(a []string) (any, error) { return lcuImport(arg(a, 0), arg(a, 1) == "true") })
	bindAsync(w, "lcuRecent", func(a []string) (any, error) { return lcuRecent() })
	bindAsync(w, "lcuMyBuilds", func(a []string) (any, error) { return lcuMyBuilds() })
	bindAsync(w, "lcuPhase", func(a []string) (any, error) { return lcuPhase() })
	bindAsync(w, "roomSeal", func(a []string) (any, error) { return roomSeal(arg(a, 0)) })
	bindAsync(w, "roomOpen", func(a []string) (any, error) { return roomOpen(arg(a, 0), arg(a, 1)) })
	if CloudKey != "" { // nuvem no Drive (só na versão privada)
		bindAsync(w, "cloudGet", func(a []string) (any, error) { return cloudGet() })
		bindAsync(w, "cloudPut", func(a []string) (any, error) { return nil, cloudPut(arg(a, 0), arg(a, 1)) })
	}
	for i, f := range extraBinds { // recursos só da versão de teste
		f(w)
		logStep(fmt.Sprintf("recurso extra %d ok", i+1))
	}
	{
		w.Bind("applyUpdate", func() (string, error) {
			exe, err := os.Executable()
			if err != nil {
				return "", err
			}
			newPath, err := stageUpdate(exe)
			if err != nil {
				return "", err
			}
			cmd := exec.Command(newPath, "--cleanup="+exe+".old")
			cmd.Dir = filepath.Dir(newPath)
			if err := cmd.Start(); err != nil {
				_ = os.Rename(exe+".old", exe)
				return "", fmt.Errorf("não consegui abrir a versão nova (%v)", err)
			}
			go func() { time.Sleep(400 * time.Millisecond); w.Dispatch(func() { w.Terminate() }) }()
			return newPath, nil
		})
	}

	w.Navigate("http://" + ln.Addr().String() + "/index.html")
	logStep("abrindo a tela")
	w.Run()
	logStep("janela fechada normalmente")
}

// chave das salas online (definida em roomkey_secret.go; sem ela o app usa uma chave de teste)
var roomSecret = ""

// lê no dados.json se o tema de Halloween está ligado (padrão: ligado)
func halloweenPref(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	var st struct {
		Opts struct {
			Halloween *bool `json:"halloween"`
		} `json:"opts"`
	}
	if json.Unmarshal(b, &st) != nil || st.Opts.Halloween == nil {
		return true
	}
	return *st.Opts.Halloween
}

// registro de abertura (inicio.log na pasta de dados): mostra até onde o programa chegou
var startLogPath string

func startLog(dir string) {
	startLogPath = filepath.Join(dir, "inicio.log")
	_ = os.WriteFile(startLogPath, nil, 0o644)
}

func logStep(s string) {
	if startLogPath == "" {
		return
	}
	f, err := os.OpenFile(startLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	fmt.Fprintf(f, "%s  %s\n", time.Now().Format("15:04:05.000"), s)
	f.Close()
}

// ---- funções em segundo plano (antes em asyncbind_windows.go) ----
// As funções ligadas com w.Bind rodam na mesma linha que desenha a janela:
// enquanto uma delas espera a internet (Drive, Google, FunLeague), a janela trava.
// bindAsync deixa a função em segundo plano e devolve o resultado depois, como uma Promise.

const asyncJS = `(function(){
  var pend = {}, seq = 0;
  window.__adone = function(id, err, res){ var p = pend[id]; delete pend[id]; if(!p) return; if(err) p.rej(new Error(err)); else p.res(res); };
  window.__amk = function(name){
    window[name] = function(){
      var args = Array.prototype.map.call(arguments, function(x){ return typeof x === "string" ? x : JSON.stringify(x); });
      return new Promise(function(res, rej){ var id = ++seq; pend[id] = {res: res, rej: rej}; window["__a_" + name](id, args); });
    };
  };
})();`

func bindAsync(w webview2.WebView, name string, fn func(args []string) (any, error)) {
	w.Bind("__a_"+name, func(id int, args []string) {
		go func() {
			var res any
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("erro interno: %v", r)
					}
				}()
				res, err = fn(args)
			}()
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			eb, _ := json.Marshal(msg)
			rb, jerr := json.Marshal(res)
			if jerr != nil {
				rb = []byte("null")
			}
			w.Dispatch(func() { w.Eval(fmt.Sprintf("window.__adone(%d,%s,%s)", id, eb, rb)) })
		}()
	})
	w.Init("window.__amk(" + mustJSON(name) + ");")
}

func mustJSON(s string) string { b, _ := json.Marshal(s); return string(b) }

func arg(a []string, i int) string {
	if i < len(a) {
		return a[i]
	}
	return ""
}
