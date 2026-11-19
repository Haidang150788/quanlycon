//go:build windows

package main

// Ban 2 (loi tat) + Ban 3 (khoa cai) — phan goi Windows API.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ---------- Ban 2: loi tat "Xin them gio" ngoai man hinh moi tai khoan ----------
func duongLoiTat() string {
	pub := os.Getenv("PUBLIC")
	if pub == "" {
		pub = `C:\Users\Public`
	}
	return filepath.Join(pub, "Desktop", "Xin thêm giờ.url")
}

func taoLoiTat() error {
	nd := "[InternetShortcut]\r\nURL=http://" + DiaChiTrangCon + "/\r\nIconIndex=0\r\nIconFile=" +
		filepath.Join(os.Getenv("ProgramFiles"), "QuanLyCon", "quanlycon.exe") + "\r\n"
	return os.WriteFile(duongLoiTat(), []byte(nd), 0644)
}

// ---------- Ban 3: khoa cai ----------
type ntWin struct{}

func nenTangThat() NenTang { return ntWin{} }

func (ntWin) QuetTienTrinh() []TienTrinh {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snap)
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	var ds []TienTrinh
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		ds = append(ds, TienTrinh{PID: e.ProcessID, Cha: e.ParentProcessID, Ten: windows.UTF16ToString(e.ExeFile[:])})
	}
	return ds
}

func (ntWin) ChiTiet(pid uint32) (ChiTietTT, error) {
	var ct ChiTietTT
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ct, err
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, 1024)
	n := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &n); err != nil {
		return ct, err
	}
	ct.DuongDan = windows.UTF16ToString(buf[:n])
	if err := windows.ProcessIdToSessionId(pid, &ct.Phien); err != nil {
		return ct, err
	}
	var tok windows.Token
	if windows.OpenProcessToken(h, windows.TOKEN_QUERY, &tok) == nil {
		ct.NangQuyen = tok.IsElevated()
		tok.Close()
	}
	return ct, nil
}

const PROCESS_SUSPEND_RESUME = 0x0800

func goiNt(p *windows.LazyProc, pid uint32) error {
	h, err := windows.OpenProcess(PROCESS_SUSPEND_RESUME, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	if r, _, _ := p.Call(uintptr(h)); r != 0 {
		return fmt.Errorf("NTSTATUS 0x%x", r)
	}
	return nil
}

func (ntWin) TamDung(pid uint32) error { return goiNt(pNtSuspendProcess, pid) }
func (ntWin) TiepTuc(pid uint32) error { return goiNt(pNtResumeProcess, pid) }
func (ntWin) Giet(pid uint32) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}
func (ntWin) PhienManHinh() uint32       { return windows.WTSGetActiveConsoleSessionId() }
func (ntWin) Bao(tieuDe, noiDung string) { baoNguoiDung(tieuDe, noiDung) }

func thuMucHeThong() []string {
	var ds []string
	for _, bien := range []string{"SystemRoot", "ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
		if v := os.Getenv(bien); v != "" {
			ds = append(ds, strings.ToLower(strings.TrimRight(v, `\`))+`\`)
		}
	}
	return ds
}

var errThay = fmt.Errorf("thay exe")

// Thu muc co it nhat mot file .exe trong vong 4 cap (dung som khi thay; toi da 5000 muc).
func coExe(goc string) bool {
	dem := 0
	sau := strings.Count(goc, `\`)
	err := filepath.WalkDir(goc, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		dem++
		if dem > 5000 {
			return filepath.SkipAll
		}
		if d.IsDir() {
			if strings.Count(p, `\`)-sau >= 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ".exe") {
			return errThay
		}
		return nil
	})
	return err == errThay
}

func quetCon(cha string, bo map[string]bool, ds *[]string) {
	muc, err := os.ReadDir(cha)
	if err != nil {
		return
	}
	for _, m := range muc {
		if !m.IsDir() || bo[strings.ToLower(m.Name())] {
			continue
		}
		p := filepath.Join(cha, m.Name())
		if coExe(p) {
			*ds = append(*ds, strings.ToLower(p)+`\`)
		}
	}
}

// Thu muc app dang co trong AppData cua moi tai khoan + ProgramData (cung cap voi thuMucApp trong khoacai.go).
func quetNenApp() []string {
	var ds []string
	users := filepath.Join(filepath.VolumeName(os.Getenv("SystemRoot"))+`\`, "Users")
	nguoi, _ := os.ReadDir(users)
	for _, n := range nguoi {
		ten := strings.ToLower(n.Name())
		if !n.IsDir() || ten == "public" || ten == "default" || ten == "default user" || ten == "all users" {
			continue
		}
		ad := filepath.Join(users, n.Name(), "AppData")
		quetCon(filepath.Join(ad, "Local"), map[string]bool{"temp": true, "packages": true, "programs": true, "microsoft": true}, &ds)
		quetCon(filepath.Join(ad, "Local", "Programs"), nil, &ds)
		quetCon(filepath.Join(ad, "Local", "Microsoft"), nil, &ds)
		quetCon(filepath.Join(ad, "Roaming"), nil, &ds)
		quetCon(filepath.Join(ad, "LocalLow"), nil, &ds)
	}
	pd := os.Getenv("ProgramData")
	if pd == "" {
		pd = `C:\ProgramData`
	}
	quetCon(pd, map[string]bool{"quanlycon": true, "microsoft": true}, &ds)
	quetCon(filepath.Join(pd, "Microsoft"), nil, &ds)
	return ds
}
