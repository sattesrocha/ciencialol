package main

import ("testing"; "os"; "net/http"; "net/http/httptest"; "bytes")

func TestNewer(t *testing.T) {
	cases := []struct{a,b string; w bool}{{"2.5","2.4",true},{"2.10","2.9",true},{"2.4","2.4",false},{"2.4","2.5",false},{"3","2.9.9",true},{"2.5.1","2.5",true}}
	for _, c := range cases { if newer(c.a,c.b)!=c.w { t.Errorf("newer(%s,%s)", c.a, c.b) } }
}
func TestDownload(t *testing.T) {
	exe := append([]byte("MZ"), bytes.Repeat([]byte{0}, 2<<20)...)
	exe = append(exe, []byte("xx CIENCIALOL_VERSION=2.7; yy")...)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ if r.URL.Path=="/html" { w.Write([]byte("<html>login</html>")); return }; w.Write(exe) }))
	defer srv.Close()
	_, v, err := download(srv.URL+"/exe"); if err!=nil || v!="2.7" { t.Fatal(v, err) }
	if _, _, err := download(srv.URL+"/html"); err==nil { t.Fatal("html deveria falhar") }
}
func TestStage(t *testing.T) {
	d := t.TempDir(); exe := d+"/CienciaLOL.exe"; os.WriteFile(exe, []byte("old"), 0o755)
	pending.data, pending.version = []byte("MZnew"), "2.6"
	np, err := stageUpdate(exe); if err!=nil { t.Fatal(err) }
	b,_ := os.ReadFile(np); if string(b)!="MZnew" || np!=exe { t.Fatal(np) }
	if _, err := os.Stat(exe+".old"); err!=nil { t.Fatal("sem .old") }
	cleanupOld(exe+".old"); if _, err := os.Stat(exe+".old"); err==nil { t.Fatal(".old não apagado") }
	if versionIn([]byte(versionTag))!=AppVersion { t.Fatal("tag") }
}

func TestManifestFlow(t *testing.T) {
	exe := append([]byte("MZ"), bytes.Repeat([]byte{0}, 2<<20)...)
	exe = append(exe, []byte("CIENCIALOL_VERSION=9.1;")...)
	man := "\ufeffCIENCIALOL-ATUALIZACAO\r\nversao=9.1\r\narquivo=1b-mrvIUL-4fsWx4AStme8S9Vxv6k0DDz\r\n\r\n(texto)"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		switch r.URL.Path { case "/man": w.Write([]byte(man)); case "/login": w.Write([]byte("<html>Fazer login</html>")); case "/f/1b-mrvIUL-4fsWx4AStme8S9Vxv6k0DDz": w.Write(exe); default: w.WriteHeader(404) } }))
	defer srv.Close()
	fu := func(id string) string { return srv.URL+"/f/"+id }
	r := checkUpdateFrom(srv.URL+"/man", fu); if r.Status!="available" || r.Version!="9.1" { t.Fatalf("%+v", r) }
	r = checkUpdateFrom(srv.URL+"/login", fu); if r.Status!="error" { t.Fatalf("%+v", r) }
	r = checkUpdateFrom(srv.URL+"/nada", fu); if r.Status!="error" { t.Fatalf("%+v", r) }
	man = "versao=" + AppVersion + "\narquivo=1b-mrvIUL-4fsWx4AStme8S9Vxv6k0DDz"
	r = checkUpdateFrom(srv.URL+"/man", fu); if r.Status!="none" { t.Fatalf("%+v", r) }
	if p := peekUpdateFrom(srv.URL+"/man"); p.Status!="none" { t.Fatalf("peek %+v", p) }
	man = "versao=99.0\narquivo=1b-mrvIUL-4fsWx4AStme8S9Vxv6k0DDz"
	if p := peekUpdateFrom(srv.URL+"/man"); p.Status!="available" || p.Version!="99.0" { t.Fatalf("peek %+v", p) }
	if p := peekUpdateFrom(srv.URL+"/nada"); p.Status!="error" { t.Fatalf("peek %+v", p) }
}
