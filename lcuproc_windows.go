//go:build windows

package main

import (
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// pasta de instalação dos LeagueClientUx.exe abertos (pela lista de processos do Windows,
// sem abrir nenhum programa extra). O lockfile dessa pasta é sempre o do cliente atual.
func lcuProcessDirs() []string {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snap)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	var dirs []string
	for err = windows.Process32First(snap, &pe); err == nil; err = windows.Process32Next(snap, &pe) {
		name := windows.UTF16ToString(pe.ExeFile[:])
		if !strings.EqualFold(name, "LeagueClientUx.exe") && !strings.EqualFold(name, "LeagueClient.exe") {
			continue
		}
		h, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pe.ProcessID)
		if e != nil {
			continue
		}
		buf := make([]uint16, 1024)
		n := uint32(len(buf))
		if windows.QueryFullProcessImageName(h, 0, &buf[0], &n) == nil {
			dirs = append(dirs, filepath.Dir(windows.UTF16ToString(buf[:n])))
		}
		windows.CloseHandle(h)
	}
	return dirs
}
