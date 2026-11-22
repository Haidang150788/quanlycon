//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const tenDichVu = "QuanLyCon"

var (
	wtsapi32              = windows.NewLazySystemDLL("wtsapi32.dll")
	pWTSSendMessage       = wtsapi32.NewProc("WTSSendMessageW")
	pWTSDisconnectSession = wtsapi32.NewProc("WTSDisconnectSession")
	dnsapi                = windows.NewLazySystemDLL("dnsapi.dll")
	pDnsFlush             = dnsapi.NewProc("DnsFlushResolverCache")
	ntdll                 = windows.NewLazySystemDLL("ntdll.dll")
	pNtSuspendProcess     = ntdll.NewProc("NtSuspendProcess")
	pNtResumeProcess      = ntdll.NewProc("NtResumeProcess")
)

func thuMucDuLieu() string {
	pd := os.Getenv("ProgramData")
	if pd == "" {
		pd = `C:\ProgramData`
	}
	return filepath.Join(pd, "QuanLyCon")
}

// ---------- hosts ----------
const dauKhoi = "# >>> QUANLYCON (tu dong, dung sua tay) >>>"
const cuoiKhoi = "# <<< QUANLYCON <<<"

func duongHosts() string {
	return filepath.Join(os.Getenv("SystemRoot"), `System32\drivers\etc\hosts`)
}

func boKhoiCu(s string) string {
	for {
		a := strings.Index(s, dauKhoi)
		if a < 0 {
			return s
		}
		b := strings.Index(s[a:], cuoiKhoi)
		if b < 0 {
			return s[:a]
		}
		s = s[:a] + strings.TrimLeft(s[a+b+len(cuoiKhoi):], "\r\n")
	}
}

func ghiHosts(mien []string) error {
	p := duongHosts()
	cu, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	s := strings.TrimRight(boKhoiCu(string(cu)), "\r\n \t")
	if len(mien) > 0 {
		var b strings.Builder
		b.WriteString(s)
		b.WriteString("\r\n\r\n" + dauKhoi + "\r\n")
		for _, m := range mien {
			b.WriteString("0.0.0.0 " + m + "\r\n")
			b.WriteString(":: " + m + "\r\n")
		}
		b.WriteString(cuoiKhoi + "\r\n")
		s = b.String()
	} else {
		s += "\r\n"
	}
	if s == string(cu) {
		return nil
	}
	if err := os.WriteFile(p, []byte(s), 0644); err != nil {
		return err
	}
	pDnsFlush.Call()
	return nil
}

// ---------- chinh sach Chrome / Edge ----------
var khoaTrinhDuyet = []struct{ duong, anDanh string }{
	{`SOFTWARE\Policies\Google\Chrome`, "IncognitoModeAvailability"},
	{`SOFTWARE\Policies\Microsoft\Edge`, "InPrivateModeAvailability"},
}

// Luon bat trong luc phan mem con cai: khong co 3 cai nay thi con lach duoc ca hosts lan URLBlocklist.
func datChongLach() error {
	for _, k := range khoaTrinhDuyet {
		key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, k.duong, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		_ = key.SetStringValue("DnsOverHttpsMode", "off") // khong cho trinh duyet tu hoi DNS qua HTTPS (lach hosts)
		_ = key.SetDWordValue(k.anDanh, 1)                // tat cua so an danh
		_ = key.SetDWordValue("BrowserGuestModeEnabled", 0)
		key.Close()
	}
	return nil
}

// Ghi danh sach dang so 1,2,3... vao <goc>\<ten>; ds rong => xoa khoa.
func ghiDanhSachSo(goc, ten string, ds []string) error {
	duong := goc + `\` + ten
	if err := registry.DeleteKey(registry.LOCAL_MACHINE, duong); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("xoa %s: %w", duong, err)
	}
	if len(ds) == 0 {
		return nil
	}
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, duong, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()
	for i, u := range ds {
		_ = key.SetStringValue(strconv.Itoa(i+1), u)
	}
	return nil
}

// Chrome va Edge co the khac nhau: YouTube chi noi cho trinh duyet co extension dang chay (kenhyt.go).
func ghiDanhSachChan(chrome, edge []string) error {
	if err := ghiDanhSachSo(khoaTrinhDuyet[0].duong, "URLBlocklist", chrome); err != nil {
		return err
	}
	return ghiDanhSachSo(khoaTrinhDuyet[1].duong, "URLBlocklist", edge)
}

// Ban 5: ep cai extension YouTube theo kenh. Co extension thi tat luon cong cu cho nha phat trien
// (DevTools xoa duoc lop che tren trang) va che do nha phat trien (nap extension tu viet de go lop che).
// Chuoi rong => go chinh sach cua trinh duyet do.
func datExtension(chrome, edge string) error {
	for i, dong := range []string{chrome, edge} {
		goc := khoaTrinhDuyet[i].duong
		ds := []string{}
		if dong != "" {
			ds = append(ds, dong)
		}
		if err := ghiDanhSachSo(goc, "ExtensionInstallForcelist", ds); err != nil {
			return err
		}
		key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, goc, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		if dong != "" {
			_ = key.SetDWordValue("DeveloperToolsAvailability", 2)
			_ = key.SetDWordValue("ExtensionDeveloperModeSettings", 1)
		} else {
			_ = key.DeleteValue("DeveloperToolsAvailability")
			_ = key.DeleteValue("ExtensionDeveloperModeSettings")
		}
		key.Close()
	}
	return nil
}

func boChinhSach() {
	for _, k := range khoaTrinhDuyet {
		_ = registry.DeleteKey(registry.LOCAL_MACHINE, k.duong+`\URLBlocklist`)
		_ = registry.DeleteKey(registry.LOCAL_MACHINE, k.duong+`\ExtensionInstallForcelist`)
		if key, err := registry.OpenKey(registry.LOCAL_MACHINE, k.duong, registry.ALL_ACCESS); err == nil {
			_ = key.DeleteValue("DnsOverHttpsMode")
			_ = key.DeleteValue(k.anDanh)
			_ = key.DeleteValue("BrowserGuestModeEnabled")
			_ = key.DeleteValue("DeveloperToolsAvailability")
			_ = key.DeleteValue("ExtensionDeveloperModeSettings")
			key.Close()
		}
	}
}

func apDungChan(d DanhSachChan) error {
	e1 := ghiHosts(d.Mien)
	e2 := ghiDanhSachChan(d.UrlChrome, d.UrlEdge)
	if e1 != nil {
		return fmt.Errorf("hosts: %w", e1)
	}
	return e2
}

// ---------- tien trinh ----------
func dongTienTrinh(ds map[string]bool) []string {
	if len(ds) == 0 {
		return nil
	}
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snap)
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	var da []string
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		ten := strings.ToLower(windows.UTF16ToString(e.ExeFile[:]))
		if !ds[ten] {
			continue
		}
		h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, e.ProcessID)
		if err != nil {
			continue
		}
		if windows.TerminateProcess(h, 1) == nil {
			da = append(da, ten)
		}
		windows.CloseHandle(h)
	}
	return da
}

// ---------- bao + khoa ----------
func baoNguoiDung(tieuDe, noiDung string) {
	sid := windows.WTSGetActiveConsoleSessionId()
	if sid == 0xFFFFFFFF {
		return
	}
	t, _ := windows.UTF16FromString(tieuDe)
	n, _ := windows.UTF16FromString(noiDung)
	var traLoi uint32
	const MB_ICONWARNING, MB_SETFOREGROUND, MB_TOPMOST = 0x30, 0x10000, 0x40000
	go pWTSSendMessage.Call(0, uintptr(sid),
		uintptr(unsafe.Pointer(&t[0])), uintptr((len(t)-1)*2),
		uintptr(unsafe.Pointer(&n[0])), uintptr((len(n)-1)*2),
		MB_ICONWARNING|MB_SETFOREGROUND|MB_TOPMOST, 50, uintptr(unsafe.Pointer(&traLoi)), 0)
}

// Ngat phien dang dung tren man hinh => Windows ve man hinh khoa, phai nhap mat khau de vao lai.
func khoaManHinh() error {
	sid := windows.WTSGetActiveConsoleSessionId()
	if sid == 0xFFFFFFFF {
		return fmt.Errorf("không có ai đang đăng nhập")
	}
	r, _, err := pWTSDisconnectSession.Call(0, uintptr(sid), 0)
	if r == 0 {
		return err
	}
	return nil
}

// ---------- dich vu ----------
func laDichVu() bool {
	ok, err := svc.IsWindowsService()
	return err == nil && ok
}

type xuLy struct{ chay func(<-chan struct{}) }

func (x xuLy) Execute(_ []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}
	dung := make(chan struct{})
	xong := make(chan struct{})
	go func() { x.chay(dung); close(xong) }()
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				if c.Cmd == svc.Stop {
					lyDoDung.Store("tay")
				} else {
					lyDoDung.Store("tat_may")
				}
				s <- svc.Status{State: svc.StopPending}
				close(dung)
				select {
				case <-xong:
				case <-time.After(15 * time.Second):
				}
				return false, 0
			}
		case <-xong:
			return false, 1 // tu chet => Windows khoi dong lai (recovery)
		}
	}
}

func chayDichVu(f func(<-chan struct{})) { _ = svc.Run(tenDichVu, xuLy{f}) }

func choCtrlC(dung chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; close(dung) }()
}

func canAdmin() error {
	if !windows.GetCurrentProcessToken().IsElevated() {
		return fmt.Errorf("phải chạy bằng quyền Administrator (chuột phải Command Prompt > Run as administrator)")
	}
	return nil
}

// Bam dup: chua co quyen Admin thi xin (hop UAC) va chay lai chinh no voi lenh "caitay".
func caiTay() {
	if windows.GetCurrentProcessToken().IsElevated() {
		caiTayDaAdmin()
		return
	}
	exe, _ := os.Executable()
	verb, _ := windows.UTF16PtrFromString("runas")
	f, _ := windows.UTF16PtrFromString(exe)
	a, _ := windows.UTF16PtrFromString("caitay")
	if err := windows.ShellExecute(0, verb, f, a, nil, windows.SW_SHOWNORMAL); err != nil {
		fmt.Println("Không xin được quyền Administrator:", err)
		fmt.Println("Cần bấm Có/Yes ở hộp hỏi quyền (máy hỏi mật khẩu thì nhập mật khẩu tài khoản Admin).")
		dungCho()
	}
}

func chepFile(tu, den string) error {
	a, err := os.Open(tu)
	if err != nil {
		return err
	}
	defer a.Close()
	b, err := os.Create(den)
	if err != nil {
		return err
	}
	if _, err := io.Copy(b, a); err != nil {
		b.Close()
		return err
	}
	return b.Close()
}

func dungVaXoaDichVu(m *mgr.Mgr) {
	s, err := m.OpenService(tenDichVu)
	if err != nil {
		return
	}
	defer s.Close()
	_, _ = s.Control(svc.Stop)
	for i := 0; i < 30; i++ {
		st, err := s.Query()
		if err != nil || st.State == svc.Stopped {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	_ = s.Delete()
}

func caiDat(token, server string) error {
	if err := canAdmin(); err != nil {
		return err
	}
	pf := os.Getenv("ProgramFiles")
	thuMucCai := filepath.Join(pf, "QuanLyCon")
	exeDich := filepath.Join(thuMucCai, "quanlycon.exe")

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	// Danh dau "dang cai lai" de dich vu cu dung ma khong bao Sep la bi dung trai phep.
	_ = os.MkdirAll(thuMucDuLieu(), 0700)
	_ = os.WriteFile(filepath.Join(thuMucDuLieu(), "dang_cai"), []byte(time.Now().String()), 0600)
	dungVaXoaDichVu(m) // cai lai / nang cap
	time.Sleep(time.Second)

	if err := os.MkdirAll(thuMucCai, 0755); err != nil {
		return err
	}
	exe, _ := os.Executable()
	if !strings.EqualFold(filepath.Clean(exe), filepath.Clean(exeDich)) {
		if err := chepFile(exe, exeDich); err != nil {
			return fmt.Errorf("chép file: %w", err)
		}
		if err := xoaTokenTrongFile(exeDich); err != nil {
			return fmt.Errorf("xoá mã trong bản cài: %w", err)
		}
	}

	// Thu muc du lieu: chi SYSTEM + Administrators doc duoc (token nam o day).
	dl := thuMucDuLieu()
	if err := os.MkdirAll(dl, 0700); err != nil {
		return err
	}
	if out, err := exec.Command("icacls", dl, "/inheritance:r", "/grant:r", "*S-1-5-18:(OI)(CI)F", "*S-1-5-32-544:(OI)(CI)F").CombinedOutput(); err != nil {
		return fmt.Errorf("icacls: %v %s", err, out)
	}
	ten, _ := os.Hostname()
	ch := CauHinhMay{Server: server, Token: token, ID: strings.ToLower(ten)}
	if err := ghiJSON(filepath.Join(dl, "may.json"), ch); err != nil {
		return err
	}

	// Thu ket noi truoc khi cai: sai khoa thi bao ngay, khong de may chay mu.
	fmt.Println("Đang thử kết nối", server, "...")
	t := &TacNhan{ch: ch}
	t.suKien("cai", "Cài đặt phần mềm phiên bản "+PhienBan)
	if err := t.hoiServer(); err != nil {
		return fmt.Errorf("không kết nối được server: %w", err)
	}
	fmt.Printf("Kết nối OK — server đang chặn %d tên miền, %d ứng dụng.\n", len(t.ban.TenMien), len(t.ban.TienTrinh))

	s, err := m.CreateService(tenDichVu, exeDich, mgr.Config{
		StartType:   mgr.StartAutomatic,
		DisplayName: "Quan ly hoc tap",
		Description: "Chặn mạng xã hội và game theo cài đặt của bố mẹ.",
	})
	if err != nil {
		return fmt.Errorf("tạo dịch vụ: %w", err)
	}
	defer s.Close()
	_ = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
	}, 3600)
	if err := s.Start(); err != nil {
		return fmt.Errorf("khởi động dịch vụ: %w", err)
	}
	fmt.Println("Dịch vụ QuanLyCon đã chạy. Nhật ký:", filepath.Join(dl, "nhatky.log"))
	if err := taoLoiTat(); err != nil {
		fmt.Println("Cảnh báo: không tạo được biểu tượng Xin thêm giờ:", err)
	} else {
		fmt.Println("Đã tạo biểu tượng \"Xin thêm giờ\" ngoài màn hình cho mọi tài khoản.")
	}
	return nil
}

func goCaiDat() error {
	if err := canAdmin(); err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	dungVaXoaDichVu(m)
	if err := ghiHosts(nil); err != nil {
		fmt.Println("Cảnh báo: không dọn được hosts:", err)
	}
	boChinhSach()
	_ = os.Remove(duongLoiTat())
	fmt.Println("Đã gỡ dịch vụ, dọn hosts và chính sách Chrome/Edge. Thư mục còn lại (xoá tay nếu muốn):")
	fmt.Println("  ", thuMucDuLieu())
	fmt.Println("  ", filepath.Join(os.Getenv("ProgramFiles"), "QuanLyCon"))
	return nil
}
